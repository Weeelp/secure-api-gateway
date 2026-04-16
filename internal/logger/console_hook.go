package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type ConsoleHook struct {
	formatter logrus.Formatter
}

func (h *ConsoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ConsoleHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	os.Stdout.Write(line)
	return nil
}
