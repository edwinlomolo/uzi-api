package internal

import (
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/getsentry/sentry-go"
	logrusSentry "github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func NewLogger() *logrus.Logger {
	// Sentry error reporting
	if isStaging() || isProd() {
		// Error level to report
		levels := []logrus.Level{
			logrus.PanicLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
		}
		hook, err := logrusSentry.New(levels, sentry.ClientOptions{
			Dsn:              config.Config.Sentry.Dsn,
			AttachStacktrace: true,
		})
		if err != nil {
			logrus.WithError(err).Fatalln("sentry hook")
		}

		log.AddHook(hook)

		defer hook.Flush(5 * time.Second)
		logrus.RegisterExitHandler(func() { hook.Flush(5 * time.Second) })
	}

	return log
}

func GetLogger() *logrus.Logger {
	return log
}

func isStaging() bool {
	return config.Config.Server.Env == "staging"
}

func isProd() bool {
	return config.Config.Server.Env == "production"
}
