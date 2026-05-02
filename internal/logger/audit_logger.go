package logger

import (
	"os"
	"time"
)

type AuditLogger struct {
	file *os.File
}

var auditLog *AuditLogger

func InitAuditLogger(filename string) error {
	var err error
	auditLog.file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (al *AuditLogger) LogBlockedRequest(method, path, reason string, ip string) {
	timestamp := time.Now().Format(time.RFC3339)
	logLine := timestamp + " | " + method + " | " + path + " | " + reason + " | " + ip + "\n"
	al.file.WriteString(logLine)
}

func CloseAuditLogger() {
	if auditLog != nil && auditLog.file != nil {
		auditLog.file.Close()
	}
}
