package infrastructure

import (
	"context"
	"os"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLoggerService() domain.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(os.Stdout)

	return &LogrusLogger{
		logger: logger,
	}
}

func (ll *LogrusLogger) Info(ctx context.Context, message string, fields ...domain.Fields) {
	ll.logger.WithFields(ll.convertFields(fields...)).Info(message)
}

func (ll *LogrusLogger) Error(ctx context.Context, message string, fields ...domain.Fields) {
	ll.logger.WithFields(ll.convertFields(fields...)).Error(message)
}

func (ll *LogrusLogger) convertFields(fields ...domain.Fields) logrus.Fields {
	result := make(logrus.Fields)
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}
