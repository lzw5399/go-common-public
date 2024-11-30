package metric

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/log"
)

var (
	// 单次请求耗时
	RequestDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_duration",
			Help: "duration of request.",
		},
		[]string{"interface", "method", "code", "service"},
	)

	// 请求次数累加
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_count",
			Help: "count of request.",
		},
		[]string{"interface", "method", "code", "service"},
	)

	// 请求次数耗时累加
	RequestDurationTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_duration_total",
			Help: "total duration of request.",
		},
		[]string{"interface", "method", "code", "service"},
	)

	// 请求耗时分布
	RequestDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_histogram",
			Help:    "duration histogram of request.",
			Buckets: []float64{10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"interface", "method", "code", "service"},
	)
)

func Init(customCollectors ...prometheus.Collector) {
	if !fconfig.DefaultConfig.OpenMonitor {
		return
	}

	log.Infof("Starting metrics monitor...")

	// 默认内置的计数器
	prometheus.MustRegister(RequestDurationGauge)
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDurationTotalCounter)
	prometheus.MustRegister(RequestDurationHistogram)

	// 注册自定义指标收集器
	for _, collector := range customCollectors {
		prometheus.MustRegister(collector)
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":"+fconfig.DefaultConfig.MonitorPort, nil)
}

func RequestDuration() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !fconfig.DefaultConfig.OpenMonitor {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "/ready" || path == "/down" || path == "/test" || path == "/" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		duration := float64(time.Since(start)) / float64(time.Second) * 1000
		path = GetSimplePath(path)
		method := c.Request.Method
		code := strconv.Itoa(c.Writer.Status())
		service := getServerName()

		RequestDurationGauge.WithLabelValues(path, method, code, service).Set(duration)
		RequestDurationTotalCounter.WithLabelValues(path, method, code, service).Add(duration)
		RequestDurationHistogram.WithLabelValues(path, method, code, service).Observe(duration)
	}
}

func RequestCount(c *gin.Context) {
	if !fconfig.DefaultConfig.OpenMonitor {
		c.Next()
		return
	}

	path := c.FullPath()
	if path == "/ready" || path == "/down" || path == "/test" || path == "/" {
		c.Next()
		return
	}

	path = GetSimplePath(path)
	method := c.Request.Method
	code := strconv.Itoa(c.Writer.Status())
	service := getServerName()

	RequestCounter.WithLabelValues(path, method, code, service).Inc()
}

func GetSimplePath(path string) string {
	// 处理路径中的数字ID
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		// 检查段是否为纯数字
		if _, err := strconv.Atoi(segment); err == nil {
			segments[i] = ":id"
			continue
		}

		// 检查UUID带了-的格式
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			segments[i] = ":uuid"
			continue
		}

		// 检查UUID不带-的格式
		if len(segment) == 32 {
			segments[i] = ":uuid"
			continue
		}

		// 雪花算法ID
		if len(segment) == 16 {
			segments[i] = ":snowflakeId"
			continue
		}

		// 小程序ID
		if strings.HasPrefix(segment, "fc") && len(segment) == 18 {
			segments[i] = ":miniAppId"
			continue
		}

		// bson.ObjectId
		if len(segment) == 24 {
			segments[i] = ":bsonId"
			continue
		}
	}

	return strings.Join(segments, "/")
}

func getServerName() string {
	name := fconfig.DefaultConfig.ServerName
	if name == "" {
		return "Unknown"
	}

	name = strings.TrimPrefix(name, "finclip-cloud-")

	return name
}
