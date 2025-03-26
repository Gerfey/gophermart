package tests

import (
	"os"

	"github.com/sirupsen/logrus"
)

func SetupTestLogging() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}
