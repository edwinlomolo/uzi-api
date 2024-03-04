package logger

import (
	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	log := logrus.New()
	return log
}
