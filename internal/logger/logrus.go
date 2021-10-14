package logger

import (
	"net/http"

	"github.com/rislah/fakes/internal/errors"
	"github.com/sirupsen/logrus"
)

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

func (l *Logger) Info(msg interface{}, fields logrus.Fields) {
	l.logr.WithFields(fields).Info(msg)
}

func (l *Logger) Error(msg interface{}, err error, fields logrus.Fields) {
	l.log(err, fields).Error(msg)
}

func (l *Logger) Warn(msg interface{}, err error, fields logrus.Fields) {
	l.log(err, fields).Warn(msg)
}

func (l *Logger) Fatal(msg interface{}, err error, fields logrus.Fields) {
	l.log(err, fields).Fatal(msg)
}

func (l *Logger) LogRequest(r *http.Request, fields logrus.Fields) {
	if fields == nil {
		fields = make(logrus.Fields)
	}

	fields["method"] = r.Method
	fields["host"] = r.Host
	fields["url"] = r.RequestURI
	fields["headers"] = r.Header
	fields["query"] = r.URL.Query()
	fields["user_agent"] = r.UserAgent()
	fields["referer"] = r.Referer()

	l.logr.WithFields(fields).Info()
}

func (l *Logger) log(err error, fields logrus.Fields) *logrus.Entry {
	if fields == nil {
		fields = make(logrus.Fields)
	}

	switch err.(type) {
	case errors.Error:
		for k, v := range err.(errors.Error).Fields() {
			fields[k] = v
		}
	}

	return l.logr.WithError(err).WithFields(fields)
}
