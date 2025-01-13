package loghooks

import (
	"fmt"

	"github.com/sirupsen/logrus"
	cls "github.com/tencentcloud/tencentcloud-cls-sdk-go"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

// TencentCLSHook implements logrus.Hook interface for CLS
type TencentCLSHook struct {
	producer *cls.AsyncProducerClient
	config   *fconfig.LogConfig
	appId    string
	env      string
}

func NewTencentCLSHook() *TencentCLSHook {
	cfg := fconfig.DefaultConfig
	producerConfig := cls.GetDefaultAsyncProducerClientConfig()
	producerConfig.Endpoint = cfg.CLSEndpoint
	producerConfig.AccessKeyID = cfg.CLSAccessKeyID
	producerConfig.AccessKeySecret = cfg.CLSAccessKeySecret
	producerConfig.TotalSizeLnBytes = cfg.CLSTotalSizeInMB * 1024 * 1024
	producerConfig.MaxSendWorkerCount = cfg.CLSMaxWorkerCount
	producerConfig.MaxBatchSize = cfg.CLSMaxBatchSize * 1024
	producerConfig.MaxBatchCount = cfg.CLSMaxBatchCount
	producerConfig.LingerMs = cfg.CLSLingerMs

	producer, err := cls.NewAsyncProducerClient(producerConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create CLS producer: %v", err))
	}

	producer.Start()

	return &TencentCLSHook{producer: producer,
		config: &cfg.LogConfig,
		appId:  cfg.ServerName,
		env:    cfg.Env,
	}
}

// Levels returns all logrus levels
func (h *TencentCLSHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

// Fire sends log to CLS
func (h *TencentCLSHook) Fire(entry *logrus.Entry) error {
	if h.producer == nil {
		return nil
	}

	fields := make(map[string]string, len(entry.Data))
	// if _, ok := entry.Data["caller"]; ok {
	// 	callerObj := entry.Data["caller"]
	// 	if callerStr, ok := callerObj.(string); ok {
	// 		fields["caller"] = callerStr
	// 	}
	// }

	// if _, ok := entry.Data["traceid"]; ok {
	// 	obj := entry.Data["traceid"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["traceid"] = str
	// 	}
	// }

	// if _, ok := entry.Data["userid"]; ok {
	// 	obj := entry.Data["userid"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["userid"] = str
	// 	}
	// }

	// if _, ok := entry.Data["memberid"]; ok {
	// 	obj := entry.Data["memberid"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["memberid"] = str
	// 	}
	// }

	// if _, ok := entry.Data["orgid"]; ok {
	// 	obj := entry.Data["orgid"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["orgid"] = str
	// 	}
	// }

	// if _, ok := entry.Data["platform"]; ok {
	// 	obj := entry.Data["platform"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["platform"] = str
	// 	}
	// }

	// if _, ok := entry.Data["isadmin"]; ok {
	// 	obj := entry.Data["isadmin"]
	// 	if bol, ok := obj.(bool); ok {
	// 		if bol {
	// 			fields["isadmin"] = "true"
	// 		} else {
	// 			fields["isadmin"] = "false"
	// 		}
	// 	}
	// }

	// if _, ok := entry.Data["lang"]; ok {
	// 	obj := entry.Data["lang"]
	// 	if str, ok := obj.(i18n.Lang); ok {
	// 		fields["lang"] = str.String()
	// 	}
	// }

	// if _, ok := entry.Data["endpoint"]; ok {
	// 	obj := entry.Data["endpoint"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["endpoint"] = str
	// 	}
	// }

	// if _, ok := entry.Data["ip"]; ok {
	// 	obj := entry.Data["ip"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["ip"] = str
	// 	}
	// }

	// if _, ok := entry.Data["file"]; ok {
	// 	obj := entry.Data["file"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["file"] = str
	// 	}
	// }

	// if _, ok := entry.Data["msg"]; ok {
	// 	obj := entry.Data["msg"]
	// 	if str, ok := obj.(string); ok {
	// 		fields["msg"] = str
	// 	}
	// }

	// Add fields from entry
	for k, v := range entry.Data {
		fields[k] = fmt.Sprintf("%v", v)
	}

	// Add the log message
	fields["message"] = entry.Message

	// Create CLS log
	clsLog := cls.NewCLSLog(entry.Time.Unix(), fields)

	// Send log asynchronously
	if err := h.producer.SendLog(h.config.CLSTopicID, clsLog, nil); err != nil {
		return fmt.Errorf("failed to send log to CLS: %v", err)
	}

	return nil
}
