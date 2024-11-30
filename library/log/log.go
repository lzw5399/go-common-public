package log

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	fcontext "github.com/lzw5399/go-common-public/library/context"
	ferrors "github.com/lzw5399/go-common-public/library/errors"
	"github.com/lzw5399/go-common-public/library/i18n"
)

var defaultLogger *Logger

func New() (*Logger, error) {
	cfg := fconfig.DefaultConfig

	l := logrus.New()
	l.Formatter = &logrus.JSONFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableTimestamp: false,
		PrettyPrint:      false,
	}
	l.Out = os.Stdout
	l.ReportCaller = true
	l.Level = getModeByStr(cfg.LogMode)
	l.SetReportCaller(false)

	if cfg.EnableFileOutput {
		// 检测路径是否存在
		_, err := os.Stat(cfg.LogFolderPath)
		if os.IsNotExist(err) { // 如果路径不存在，则通过MkdirAll创建目录，并赋予读写权限
			err = os.MkdirAll(cfg.LogFolderPath, 0755)
			if err != nil {
				panic("failed to create directory")
			}
		} else if err != nil { // 如果返回了其他错误，则输出错误信息
			fmt.Printf("failed to determine if directory exists: %s\n", err.Error())
			panic("failed to determine if directory exists")
		}

		getFileName := func() string {
			envHostName := os.Getenv("HOSTNAME")
			serverPrefix := strings.TrimPrefix(cfg.ServerName, "finclip-cloud-") // 兜底的服务名
			if !strings.HasPrefix(envHostName, serverPrefix) {
				envHostName = serverPrefix
			}

			currentDate := time.Now().Format("2006-01-02")
			logFile := fmt.Sprintf("%s_%s.log", envHostName, currentDate)
			return path.Join(cfg.LogFolderPath, logFile)
		}

		multiWriter := io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Filename:   getFileName(),      // 日志文件位置
			MaxSize:    cfg.LogMaxSize,     // 单文件最大容量,单位是MB
			MaxBackups: cfg.LogMaxBackups,  // 最大保留过期文件个数
			MaxAge:     cfg.LogMaxSaveTime, // 保留过期文件的最大时间间隔,单位是天
			Compress:   cfg.LogCompress,    // 是否需要压缩滚动日志, 使用的 gzip 压缩
			LocalTime:  true,
		})

		l.SetOutput(multiWriter)
	}

	logger := &Logger{
		appId:            cfg.ServerName,
		env:              cfg.Env,
		enableFileOutput: cfg.EnableFileOutput,
		level:            cfg.LogMode,
		parseSvrRspInfoAndDowngrade400SerialError: cfg.ParseSvrRspInfoAndDowngrade400SerialError,
		Logger: l,
	}

	return logger, nil
}

func InitLogger() {
	var err error
	defaultLogger, err = New()
	if err != nil {
		panic(err)
	}
}

func getModeByStr(s string) logrus.Level {
	switch s {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "trace":
		return logrus.TraceLevel
	case "fatal":
		return logrus.FatalLevel
	default:
		return logrus.DebugLevel
	}
}

type Logger struct {
	appId                                     string
	env                                       string
	enableFileOutput                          bool
	level                                     string
	parseSvrRspInfoAndDowngrade400SerialError bool
	*logrus.Logger
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	tracef(newEntryWithEnv(l), format, args...)
}

func (l *Logger) Tracec(ctx context.Context, format string, args ...interface{}) {
	tracec(ctx, newEntryWithEnv(l), format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	debugf(newEntryWithEnv(l), format, args...)
}

func (l *Logger) Debugc(ctx context.Context, format string, args ...interface{}) {
	debugc(ctx, newEntryWithEnv(l), format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	infof(newEntryWithEnv(l), format, args...)
}

func (l *Logger) Infoc(ctx context.Context, format string, args ...interface{}) {
	infoc(ctx, newEntryWithEnv(l), format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	warnf(newEntryWithEnv(l), format, args...)
}

func (l *Logger) Warnc(ctx context.Context, format string, args ...interface{}) {
	warnc(ctx, newEntryWithEnv(l), format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.parseSvrRspInfoAndDowngrade400SerialError && checkShouldDowngrade(args...) {
		warnf(newEntryWithEnv(l), format, args...)
		return
	}
	errorf(newEntryWithEnv(l), format, args...)
}

func (l *Logger) Errorc(ctx context.Context, format string, args ...interface{}) {
	errorc(ctx, newEntryWithEnv(l), format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	fatalf(newEntryWithEnv(l), format, args...)
}

func (l *Logger) Fatalc(ctx context.Context, format string, args ...interface{}) {
	fatalc(ctx, newEntryWithEnv(l), format, args...)
}

func (l *Logger) WithField(key string, value interface{}) *Entry {
	entry := newEntryWithEnv(l)
	return entry.WithField(key, value)
}

func (l *Logger) WithFields(fields logrus.Fields) *Entry {
	entry := newEntryWithEnv(l)
	return entry.WithFields(fields)
}

func Tracef(format string, args ...interface{}) {
	tracef(newEntryWithEnv(defaultLogger), format, args...)
}

func Tracec(ctx context.Context, format string, args ...interface{}) {
	tracec(ctx, newEntryWithEnv(defaultLogger), format, args...)
}

func Debugf(format string, args ...interface{}) {
	debugf(newEntryWithEnv(defaultLogger), format, args...)
}

func Debugc(ctx context.Context, format string, args ...interface{}) {
	debugc(ctx, newEntryWithEnv(defaultLogger), format, args...)
}

func Infof(format string, args ...interface{}) {
	infof(newEntryWithEnv(defaultLogger), format, args...)
}

func Infoc(ctx context.Context, format string, args ...interface{}) {
	infoc(ctx, newEntryWithEnv(defaultLogger), format, args...)
}

func Warnf(format string, args ...interface{}) {
	warnf(newEntryWithEnv(defaultLogger), format, args...)
}

func Warnc(ctx context.Context, format string, args ...interface{}) {
	warnc(ctx, newEntryWithEnv(defaultLogger), format, args...)
}

func Errorf(format string, args ...interface{}) {
	errorf(newEntryWithEnv(defaultLogger), format, args...)
}

func Errorc(ctx context.Context, format string, args ...interface{}) {
	errorc(ctx, newEntryWithEnv(defaultLogger), format, args...)
}

func Fatalf(format string, args ...interface{}) {
	fatalf(newEntryWithEnv(defaultLogger), format, args...)
}

func Fatalc(ctx context.Context, format string, args ...interface{}) {
	fatalc(ctx, newEntryWithEnv(defaultLogger), format, args...)
}

func appendContextFields(ctx context.Context, entry *Entry) *Entry {
	if ctx == nil {
		return entry
	}

	// caller
	caller := fcontext.CallerFromContext(ctx)
	if caller != "" {
		entry = entry.WithField("caller", caller)
	}

	// trace id
	traceId := fcontext.TraceIdFromContext(ctx)
	if traceId != "" {
		entry = entry.WithField("traceid", traceId)
	}

	// user info
	userInfo := fcontext.UserInfoFromContext(ctx)
	if userInfo != nil {
		entry = entry.WithField("accountid", userInfo.AccountId)
		entry = entry.WithField("memberid", userInfo.MemberId)
		entry = entry.WithField("organid", userInfo.OrganId)
		entry = entry.WithField("platform", userInfo.PlatForm)
		entry = entry.WithField("isadmin", userInfo.IsAdmin)
	}

	// lang
	lang := fcontext.LangFromContext(ctx)
	entry = entry.WithField(i18n.HeaderLang, lang)

	// http endpoint
	endpoint := fcontext.HttpEndpointFromContext(ctx)
	if endpoint != "" {
		entry = entry.WithField("endpoint", endpoint)
	}

	// ip
	ip := fcontext.ClientIpFromContext(ctx)
	if ip != "" {
		entry = entry.WithField("ip", ip)
	}

	return entry
}

func appendRuntimeFields(entry *Entry) *Entry {
	_, fileName, lineNum, ok := runtime.Caller(3)
	if !ok {
		return entry
	}

	filePath := filepath.Dir(fileName)
	file := filepath.Base(fileName)
	entry = entry.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s/%s:%d", filePath, file, lineNum),
	})
	return entry
}

func tracef(entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.TraceLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry.Entry.Tracef(format, args...)
}

func tracec(ctx context.Context, entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.TraceLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry = appendContextFields(ctx, entry)
	entry.Entry.Tracef(format, args...)
}

func debugf(entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.DebugLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry.Entry.Debugf(format, args...)
}

func debugc(ctx context.Context, entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.DebugLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry = appendContextFields(ctx, entry)
	entry.Entry.Debugf(format, args...)
}

func infof(entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.InfoLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry.Entry.Infof(format, args...)
}

func infoc(ctx context.Context, entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.InfoLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry = appendContextFields(ctx, entry)
	entry.Entry.Infof(format, args...)
}

func warnf(entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.WarnLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry.Entry.Warnf(format, args...)
}

func warnc(ctx context.Context, entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.WarnLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry = appendContextFields(ctx, entry)
	entry.Entry.Warnf(format, args...)
}

func errorf(entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.ErrorLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry.Entry.Errorf(format, args...)
}

func errorc(ctx context.Context, entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.ErrorLevel {
		return
	}
	if defaultLogger.parseSvrRspInfoAndDowngrade400SerialError && checkShouldDowngrade(args...) {
		warnc(ctx, entry, format, args...)
		return
	}
	entry = appendRuntimeFields(entry)
	entry = appendContextFields(ctx, entry)
	entry.Entry.Errorf(format, args...)
}

func fatalf(entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.FatalLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry.Entry.Fatalf(format, args...)
}

func fatalc(ctx context.Context, entry *Entry, format string, args ...interface{}) {
	if entry.Level < logrus.FatalLevel {
		return
	}
	entry = appendRuntimeFields(entry)
	entry = appendContextFields(ctx, entry)
	entry.Entry.Fatalf(format, args...)
}

func checkShouldDowngrade(args ...interface{}) bool {
	shouldDowngrade := false
	for _, arg := range args {
		switch argType := arg.(type) {
		case *ferrors.SvrRspInfo:
			// 400 <= argType.HttpStatus < 500
			if argType.HttpStatus >= http.StatusBadRequest && argType.HttpStatus < http.StatusInternalServerError {
				shouldDowngrade = true
			}
		case error:
			rspInfo := ferrors.ExtractSvrRspInfo(argType)
			// 400 <= argType.HttpStatus < 500
			if rspInfo.HttpStatus >= http.StatusBadRequest && rspInfo.HttpStatus < http.StatusInternalServerError {
				shouldDowngrade = true
			}
		}
	}
	return shouldDowngrade
}

func newEntryWithEnv(l *Logger) *Entry {
	entry := l.Logger.WithField("appid", l.appId).WithField("env", l.env)
	entry.Level = l.Logger.Level
	return &Entry{
		Entry: entry,
		parseSvrRspInfoAndDowngrade400SerialError: l.parseSvrRspInfoAndDowngrade400SerialError,
	}
}
