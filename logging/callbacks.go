package logging

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	callbacks2 "github.com/cloudwego/eino/utils/callbacks"
	"github.com/sirupsen/logrus"
	"github.com/thalesfu/golangutils/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

func GetGraphTraceCallBacksOpts() compose.Option {
	return compose.WithCallbacks(callbacks2.NewHandlerHelper().Graph(GetCallBackHandlers()).Handler())
}

func GetCallBackHandlers() callbacks.Handler {
	return callbacks.NewHandlerBuilder().OnStartFn(onStart).OnEndFn(onEnd).OnErrorFn(onError).Build()
}

func onStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	now := time.Now()

	opts := []oteltrace.SpanStartOption{
		oteltrace.WithTimestamp(now),
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
	}

	spanName := fmt.Sprintf("%s-%s", info.Component, info.Name)

	parentSpan := oteltrace.SpanFromContext(ctx)
	tracer := otel.Tracer(parentSpan.SpanContext().TraceID().String())
	ctx, _ = tracer.Start(ctx, spanName, opts...)
	ctx, logStore := logging.InitializeContextLogStore(ctx, spanName)

	logStore.Set("in", fmt.Sprint(input))
	logStore.Set("Runnable.Component", string(info.Component))
	logStore.Set("Runnable.Type", info.Type)
	logStore.Set("Runnable.Name", info.Name)

	return ctx
}

func onEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	currentSpan := oteltrace.SpanFromContext(ctx)
	logStore, ok := logging.GetContextLogStore(ctx)
	if ok {
		logStore.Set("out", fmt.Sprint(output))

		logData := logStore.GetAll()
		for k, v := range logData {
			currentSpan.SetAttributes(attribute.String(k, v))
		}
	}

	currentSpan.SetStatus(codes.Ok, "")
	currentSpan.End()

	logrus.WithContext(ctx).Infof("Node process Success: name: %s, type: %s, component: %s ", info.Name, info.Type, info.Component)

	return ctx
}

func onError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	currentSpan := oteltrace.SpanFromContext(ctx)
	logStore, ok := logging.GetContextLogStore(ctx)
	if ok {
		logStore.Set("error", err.Error())

		logData := logStore.GetAll()
		for k, v := range logData {
			currentSpan.SetAttributes(attribute.String(k, v))
		}
	}

	currentSpan.SetStatus(codes.Error, err.Error())
	currentSpan.End()

	logrus.WithContext(ctx).Errorf("Node process Error: name: %s, type: %s, component: %s ", info.Name, info.Type, info.Component)

	return ctx
}
