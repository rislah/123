package logger

import (
	"net/http"

	"github.com/rislah/fakes/internal/errors"
	"github.com/sirupsen/logrus"
)

var SharedGlobalLogger = New("development")

type Logger struct {
	logr *logrus.Entry
}

func New(env string) *Logger {
	log := logrus.New()

	switch env {
	case "development":
		log.Formatter = &logrus.TextFormatter{}
		log.Level = logrus.DebugLevel
	}

	fields := logrus.Fields{
		"env": env,
	}

	logWithFields := log.WithFields(fields)

	return &Logger{
		logWithFields,
	}
}

func (l *Logger) InfoWithFields(msg interface{}, fields logrus.Fields) {
	l.logr.WithFields(fields).Info(msg)
}

func (l *Logger) ErrorWithFields(msg interface{}, err error, fields logrus.Fields) {
	l.log(err, fields).Error(msg)
}

func (l *Logger) WarnWithFields(msg interface{}, err error, fields logrus.Fields) {
	l.log(err, fields).Warn(msg)
}

func (l *Logger) FatalWithFields(msg interface{}, err error, fields logrus.Fields) {
	l.log(err, fields).Fatal(msg)
}

func (l *Logger) Info(msg interface{}) {
	l.logr.Info(msg)
}

func (l *Logger) Error(msg interface{}, err error) {
	l.log(err).Error(msg)
}

func (l *Logger) Warn(msg interface{}, err error) {
	l.log(err).Warn(msg)
}

func (l *Logger) Fatal(msg interface{}, err error) {
	l.log(err).Fatal(msg)
}

func (l *Logger) LogRequest(r *http.Request, fields ...logrus.Fields) {
	mergeFields := make(logrus.Fields)

	mergeFields["method"] = r.Method
	mergeFields["host"] = r.Host
	mergeFields["url"] = r.RequestURI
	mergeFields["headers"] = r.Header
	mergeFields["query"] = r.URL.Query()
	mergeFields["user_agent"] = r.UserAgent()
	mergeFields["referer"] = r.Referer()

	if fields != nil {
		for k, v := range fields[0] {
			mergeFields[k] = v
		}
	}

	l.logr.WithFields(mergeFields).Info()
}

func (l *Logger) LogRequestError(err error, r *http.Request, fields ...logrus.Fields) {
	mergeFields := make(logrus.Fields)

	mergeFields["method"] = r.Method
	mergeFields["host"] = r.Host
	mergeFields["url"] = r.RequestURI
	mergeFields["headers"] = r.Header
	mergeFields["query"] = r.URL.Query()
	mergeFields["user_agent"] = r.UserAgent()
	mergeFields["referer"] = r.Referer()

	if fields != nil {
		for k, v := range fields[0] {
			mergeFields[k] = v
		}
	}

	l.log(err).WithFields(mergeFields).Error()
}

func (l *Logger) log(err error, fields ...logrus.Fields) *logrus.Entry {
	mergeFields := make(logrus.Fields)

	switch err.(type) {
	case errors.Error:
		for k, v := range err.(errors.Error).Fields() {
			mergeFields[k] = v
		}
	}

	if fields != nil {
		for k, v := range fields[0] {
			mergeFields[k] = v
		}
	}

	return l.logr.WithError(err).WithFields(mergeFields)
}
