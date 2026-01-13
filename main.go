package main

import (
	"context"
	"time"

	"github.com/phoenix/golang-logging/internal/config"
	"github.com/phoenix/golang-logging/internal/logger"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	ctx := context.Background()
	shutdown, err := logger.InitOTLP(ctx, cfg)
	if err != nil {
		panic("failed to initialize OTLP logger: " + err.Error())
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			logger.Logger.Error("failed to shutdown OTLP", zap.Error(err))
		}
	}()

	// Create a tracer
	tracer := otel.Tracer("golang-logging-demo")

	// Start a span to demonstrate trace context
	ctx, span := tracer.Start(ctx, "main-operation")
	defer span.End()

	logger.InfoContext(ctx, "Application started")

	// Create a child span for database operation
	ctx, dbSpan := tracer.Start(ctx, "database-query")

	logger.WarnContext(ctx, "Warning message",
		zap.String("resource", "database"),
		zap.Float64("usage_percent", 85.5))

	logger.ErrorContext(ctx, "Error occurred",
		zap.String("error_code", "DB_CONNECTION_FAILED"),
		zap.Int("retry_count", 3),
		zap.String("severity", "high"))

	dbSpan.End()

	// Create another span for HTTP request
	ctx, httpSpan := tracer.Start(ctx, "http-request")

	logger.InfoContext(ctx, "Processing request",
		zap.String("request_id", "req-abc-123"),
		zap.String("user_agent", "Mozilla/5.0"),
		zap.String("ip", "192.168.1.100"),
		zap.Int("duration_ms", 245))

	httpSpan.End()

	time.Sleep(2 * time.Second)

	logger.InfoContext(ctx, "Application shutting down")
}
