package log

import (
	"context"
	"fmt"
	"strconv"
	"time"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/pkg/errors"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type GormJsonLogger struct {
	l                                   *Logger
	gormConfig                          *gormLogger.Config
	level                               gormLogger.LogLevel
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewGormJsonLogger(gormConfig *gormLogger.Config) *GormJsonLogger {
	cfg := fconfig.DefaultConfig

	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if gormConfig == nil {
		gormConfig = &gormLogger.Config{
			SlowThreshold:             time.Millisecond * time.Duration(fconfig.DefaultConfig.SlowSqlMillSeconds),
			LogLevel:                  GormLogModeStrToGormLogLevel(cfg.GormLogMode),
			IgnoreRecordNotFoundError: false,
		}
	}

	return &GormJsonLogger{
		l:            defaultLogger,
		gormConfig:   gormConfig,
		level:        GormLogModeStrToGormLogLevel(cfg.GormLogMode),
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceErrStr:  traceWarnStr,
		traceWarnStr: traceErrStr,
	}
}

func (g *GormJsonLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *g
	newLogger.level = level
	return &newLogger
}

func (g *GormJsonLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= gormLogger.Info {
		g.l.Infoc(ctx, g.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *GormJsonLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= gormLogger.Warn {
		g.l.Warnc(ctx, g.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *GormJsonLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= gormLogger.Error {
		g.l.Errorc(ctx, g.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (g *GormJsonLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	// 错误日志 & 没找到记录的日志
	case err != nil && g.level >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !g.gormConfig.IgnoreRecordNotFoundError):
		sql, rows := fc()
		rowsStr := "-"
		if rows != -1 {
			rowsStr = strconv.Itoa(int(rows))
		}
		g.l.Warnc(ctx, g.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rowsStr, sql)

	// 慢sql
	case elapsed > g.gormConfig.SlowThreshold && g.gormConfig.SlowThreshold != 0 && g.level >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", g.gormConfig.SlowThreshold)

		rowsStr := "-"
		if rows != -1 {
			rowsStr = strconv.Itoa(int(rows))
		}
		g.l.
			WithField("t", "slow_sql").
			WithField("elapsed", float64(elapsed.Nanoseconds())/1e6).
			WithField("sql", sql).
			Warnc(ctx, g.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rowsStr, sql)

	// 剩余所有日志
	case g.level == gormLogger.Info:
		sql, rows := fc()
		rowsStr := "-"
		if rows != -1 {
			rowsStr = strconv.Itoa(int(rows))
		}

		g.l.Infoc(ctx, g.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rowsStr, sql)
	}
}

func GormLogModeStrToGormLogLevel(level string) gormLogger.LogLevel {
	switch level {
	case "silent":
		return gormLogger.Silent
	case "error":
		return gormLogger.Error
	case "warn":
		return gormLogger.Warn
	case "info":
		return gormLogger.Info
	default:
		return gormLogger.Error
	}
}
