package config

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	TargetDir string
	Port      int
	Host      string
	Log       *logrus.Logger
}
