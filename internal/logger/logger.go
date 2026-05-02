package logger

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type logrusWrapper struct {
	entry logrus.FieldLogger
}

var (
	Log            Logger
	logFile        *os.File
	bufferedWriter *bufio.Writer
)

func Init(fileName string) {
	tempLog := logrus.New()

	tempLog.SetFormatter(&logrus.JSONFormatter{})
	tempLog.SetLevel(logrus.DebugLevel)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		fmt.Printf("Log file error: %v", err)
		tempLog.SetOutput(os.Stdout)
	} else {
		logFile = file
		bufferedWriter = bufio.NewWriter(file)

		tempLog.SetOutput(bufferedWriter)

		tempLog.AddHook(&ConsoleHook{
			formatter: &logrus.TextFormatter{
				FullTimestamp: true,
				ForceColors:   true,
			},
		})

		go func() {
			for {
				time.Sleep(time.Second * 1)
				if bufferedWriter != nil {
					bufferedWriter.Flush()
				}
			}
		}()
	}

	Log = &logrusWrapper{entry: tempLog}
	tempLog.Info("Success logger init")
}

func (w *logrusWrapper) argsToFields(args ...any) logrus.Fields {
	fields := make(logrus.Fields)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				fields[key] = args[i+1]
			}
		}
	}
	return fields
}

func (w *logrusWrapper) With(args ...any) Logger {
	return &logrusWrapper{
		entry: w.entry.WithFields(w.argsToFields(args...)),
	}
}

func (w *logrusWrapper) write(level logrus.Level, msg string, args ...any) {
	w.entry.WithFields(w.argsToFields(args...)).Log(level, msg)
}

func (w *logrusWrapper) Info(msg string, args ...any)  { w.write(logrus.InfoLevel, msg, args...) }
func (w *logrusWrapper) Error(msg string, args ...any) { w.write(logrus.ErrorLevel, msg, args...) }
func (w *logrusWrapper) Debug(msg string, args ...any) { w.write(logrus.DebugLevel, msg, args...) }
func (w *logrusWrapper) Warn(msg string, args ...any)  { w.write(logrus.WarnLevel, msg, args...) }
func (w *logrusWrapper) Fatal(msg string, args ...any) { w.write(logrus.FatalLevel, msg, args...) }

func Close() {
	if bufferedWriter != nil {
		bufferedWriter.Flush()
	}
	if logFile != nil {
		logFile.Close()
	}
}
