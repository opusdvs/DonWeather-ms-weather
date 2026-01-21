package domain

import "context"

type Logger interface {
	Info(ctx context.Context, message string, fields ...Fields)
	Error(ctx context.Context, message string, fields ...Fields)
}

type Fields struct {
	Key   string
	Value any
}
