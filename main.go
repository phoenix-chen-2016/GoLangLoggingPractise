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

	log := logger.Logger

	// Create a tracer
	tracer := otel.Tracer("golang-logging-demo")

	// Start a span to demonstrate trace context
	ctx, span := tracer.Start(ctx, "main-operation")
	defer span.End()

	// Get trace and span IDs from context
	spanCtx := span.SpanContext()
	log.Info("Application started",
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()))

	// Create a child span for database operation
	ctx, dbSpan := tracer.Start(ctx, "database-query")
	dbSpanCtx := dbSpan.SpanContext()
	
	log.Warn("Warning message",
		zap.String("trace_id", dbSpanCtx.TraceID().String()),
		zap.String("span_id", dbSpanCtx.SpanID().String()),
		zap.String("resource", "database"),
		zap.Float64("usage_percent", 85.5))

	log.Error("Error occurred",
		zap.String("trace_id", dbSpanCtx.TraceID().String()),
		zap.String("span_id", dbSpanCtx.SpanID().String()),
		zap.String("error_code", "DB_CONNECTION_FAILED"),
		zap.Int("retry_count", 3),
		zap.String("severity", "high"))
	
	dbSpan.End()

	// Create another span for HTTP request
	ctx, httpSpan := tracer.Start(ctx, "http-request")
	httpSpanCtx := httpSpan.SpanContext()
	
	log.Info("Processing request",
		zap.String("trace_id", httpSpanCtx.TraceID().String()),
		zap.String("span_id", httpSpanCtx.SpanID().String()),
		zap.String("request_id", "req-abc-123"),
		zap.String("user_agent", "Mozilla/5.0"),
		zap.String("ip", "192.168.1.100"),
		zap.Int("duration_ms", 245))
	
	httpSpan.End()

	time.Sleep(2 * time.Second)
	
	log.Info("Application shutting down",
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()))
}
