package server

import (
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

func SetLogger(log *logrus.Logger) {
	logger = log
	logger.Info("update controller init")
}
