package utils

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ContextLogger struct {
	logger *log.Logger
	fields log.Fields
}

type Fields log.Fields

// NewFields creates a new Fields type
func NewFields() Fields {
	return Fields{}
}

// NewContextLogger creates a new ContextLogger
func NewContextLogger(exec *ExecutionContext) *ContextLogger {
	logger := &log.Logger{
		Out:       os.Stdout,
		Formatter: new(log.JSONFormatter),
		Level:     log.GetLevel(),
	}

	fields := log.Fields{
		"awsRegion": exec.AWSRegion,
		"tenantID":  exec.TenantID,
		"firstName": exec.FirstName,
		"lastName":  exec.LastName,
		"email":     exec.Email,
		"userRole":  exec.UserRole,
		"userID":    exec.UserID,
	}

	return &ContextLogger{
		logger: logger,
		fields: fields,
	}
}

func (c *ContextLogger) ErrorLog(message string, fields Fields) {
	c.log(log.ErrorLevel, message, log.Fields(fields))
}

func (c *ContextLogger) WarnLog(message string, fields Fields) {
	c.log(log.WarnLevel, message, log.Fields(fields))
}

func (c *ContextLogger) InfoLog(message string, fields Fields) {
	c.log(log.InfoLevel, message, log.Fields(fields))
}

func (c *ContextLogger) log(level log.Level, message string, fields log.Fields) {
	// Merge shared fields with provided fields
	for k, v := range c.fields {
		fields[k] = v
	}
	entry := c.logger.WithFields(fields)
	switch level {
	case log.InfoLevel:
		entry.Info(message)
	case log.WarnLevel:
		entry.Warn(message)
	case log.ErrorLevel:
		entry.Error(message)
	}
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	logLevelEnv := os.Getenv("LOG_LEVEL")
	if logLevelEnv == "" {
		logLevelEnv = "info" // Default level
	}

	level, err := log.ParseLevel(strings.ToLower(logLevelEnv))
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}

	log.SetLevel(level)
}
