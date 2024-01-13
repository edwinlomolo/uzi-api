package logger

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func NewLogger() *logrus.Logger {
	log := logrus.New()
	logger = log

	return logger
}

func GetLogger() *logrus.Logger {
	return logger
}
