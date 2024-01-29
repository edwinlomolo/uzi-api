package logger

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func NewLogger() {
	logger = logrus.New()
}

func GetLogger() *logrus.Logger {
	return logger
}
