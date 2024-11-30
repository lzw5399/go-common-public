package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Entry struct {
	*logrus.Entry
	parseSvrRspInfoAndDowngrade400SerialError bool
}

func WithField(key string, value interface{}) *Entry {
	entry := newEntryWithEnv(defaultLogger)
	return entry.WithField(key, value)
}

func WithFields(fields logrus.Fields) *Entry {
	entry := newEntryWithEnv(defaultLogger)
	return entry.WithFields(fields)
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	e.Entry = e.Entry.WithFields(logrus.Fields{key: value})
	return e
}

func (e *Entry) WithFields(fields logrus.Fields) *Entry {
	e.Entry = e.Entry.WithFields(fields)
	return e
}

func (e *Entry) Tracef(format string, args ...interface{}) {
	tracef(e, format, args...)
}

func (e *Entry) Tracec(ctx context.Context, format string, args ...interface{}) {
	tracec(ctx, e, format, args...)
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	debugf(e, format, args...)
}

func (e *Entry) Debugc(ctx context.Context, format string, args ...interface{}) {
	debugc(ctx, e, format, args...)
}

func (e *Entry) Infof(format string, args ...interface{}) {
	infof(e, format, args...)
}

func (e *Entry) Infoc(ctx context.Context, format string, args ...interface{}) {
	infoc(ctx, e, format, args...)
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	warnf(e, format, args...)
}

func (e *Entry) Warnc(ctx context.Context, format string, args ...interface{}) {
	warnc(ctx, e, format, args...)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	if e.parseSvrRspInfoAndDowngrade400SerialError && checkShouldDowngrade(args...) {
		warnf(e, format, args...)
		return
	}
	errorf(e, format, args...)
}

func (e *Entry) Errorc(ctx context.Context, format string, args ...interface{}) {
	errorc(ctx, e, format, args...)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	fatalf(e, format, args...)
}

func (e *Entry) Fatalc(ctx context.Context, format string, args ...interface{}) {
	fatalc(ctx, e, format, args...)
}
