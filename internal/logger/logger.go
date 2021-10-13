package logger

import (
	"fmt"
	"net/http"

	"github.com/rislah/fakes/internal/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(env string) *Logger {
	var logger *zap.Logger
	opts := zap.Fields(zap.String("env", env))

	switch env {
	case "development":
		logger, _ = zap.NewDevelopment(opts, zap.AddStacktrace(zapcore.FatalLevel))
	}

	_ = zap.ReplaceGlobals(logger)

	return &Logger{logger}
}

func (l *Logger) LogWarn(err error, msg string, zapFields ...zap.Field) {
	l.log(err, msg, zap.WarnLevel, zapFields...)
}

func (l *Logger) LogError(err error, msg string, zapFields ...zap.Field) {
	l.log(err, msg, zap.ErrorLevel, zapFields...)
}

func (l *Logger) log(err error, msg string, logLevel zapcore.Level, zapFields ...zap.Field) {
	var fields []zap.Field

	switch err.(type) {
	case errors.Error:
		fields = l.fieldsFromError(err.(errors.Error))
	default:
	}

	fields = append(fields, zapFields...)
	log := l.Logger.With(fields...)

	switch logLevel {
	case zap.WarnLevel:
		log.Warn(msg, zap.Error(err))
	case zap.ErrorLevel:
		log.Error(msg, zap.Error(err))
	}
}

func (l *Logger) LogRequestError(r *http.Request, err error, msg string, zapFields ...zap.Field) {
	l.LogError(err, msg, zap.Object("request", logRequest{r}))
}

func (l *Logger) fieldsFromError(errx errors.Error) []zap.Field {
	var fields []zap.Field
	for k, v := range errx.Fields() {
		fields = append(fields, zap.Any(k, v))
	}

	return fields
}

func (l *Logger) LogRequest(req *http.Request, err error) {
	l.Logger.Error(fmt.Sprintf("%s %s", req.Method, req.URL.String()), zap.Error(err), zap.Object("request", logRequest{req}))
}

type logRequest struct {
	*http.Request
}

func (lq logRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("host", lq.Host)
	enc.AddString("method", lq.Method)
	enc.AddString("uri", lq.RequestURI)
	_ = enc.AddObject("headers", multimap(lq.Header))
	_ = enc.AddObject("query", multimap(lq.URL.Query()))
	enc.AddString("user_agent", lq.UserAgent())
	enc.AddString("referer", lq.Referer())
	return nil
}

type multimap map[string][]string

func (m multimap) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for k, v := range m {
		zap.Strings(k, v).AddTo(encoder)
	}
	return nil
}
