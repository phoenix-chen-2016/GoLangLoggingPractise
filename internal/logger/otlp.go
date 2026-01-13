package logger

import (
	"context"
	"log"
	"os"

	"github.com/phoenix/golang-logging/internal/config"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Logger *zap.Logger

func InitOTLP(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.Service.Name),
			semconv.ServiceVersion(cfg.Service.Version),
		),
	)
	if err != nil {
		return nil, err
	}

	var grpcOpts []grpc.DialOption
	if cfg.OTLP.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Initialize Trace Exporter
	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLP.Endpoint),
	}
	if cfg.OTLP.Insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
		traceOpts = append(traceOpts, otlptracegrpc.WithDialOption(grpcOpts...))
	}

	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set propagator for trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize Log Exporter
	logOpts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.OTLP.Endpoint),
	}
	if cfg.OTLP.Insecure {
		logOpts = append(logOpts, otlploggrpc.WithInsecure())
		logOpts = append(logOpts, otlploggrpc.WithDialOption(grpcOpts...))
	}

	logExporter, err := otlploggrpc.New(ctx, logOpts...)
	if err != nil {
		return nil, err
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	global.SetLoggerProvider(loggerProvider)

	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:       zapcore.OmitKey,
		LevelKey:      "level",
		NameKey:       zapcore.OmitKey,
		CallerKey:     zapcore.OmitKey,
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: zapcore.OmitKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
	}

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	otelCore := otelzap.NewCore(
		cfg.Service.Name,
		otelzap.WithLoggerProvider(loggerProvider),
	)

	core := zapcore.NewTee(consoleCore, otelCore)

	zapLogger := zap.New(core, zap.AddCaller())

	Logger = zapLogger.With(
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
	)

	log.Println("✅ OTLP Logger initialized successfully, sending logs to:", cfg.OTLP.Endpoint)
	log.Println("✅ OTLP Tracer initialized successfully, trace context will be included in logs")

	shutdown := func(ctx context.Context) error {
		Logger.Sync()
		if err := loggerProvider.Shutdown(ctx); err != nil {
			return err
		}
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}

	return shutdown, nil
}
