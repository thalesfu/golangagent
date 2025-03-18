package mem

import (
	"context"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	callbacks2 "github.com/cloudwego/eino/utils/callbacks"
)

const SessionContextKey = "ctx-session"

func GetInitMemGraphCallBackOption() compose.Option {
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(graphInitMem).
		OnStartFn(graphInitMem).
		Build()

	return compose.WithCallbacks(callbacks2.NewHandlerHelper().Graph(handler).Handler())
}

func graphInitMem(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	session := NewSession("", 0, "", "", "", nil)
	ctx = context.WithValue(ctx, SessionContextKey, session)
	return ctx
}

func GetSessionFromContext(ctx context.Context) *Session {
	return ctx.Value(SessionContextKey).(*Session)
}

func GetModelMemCallBackOption(nodeKey string) compose.Option {
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(setInputMessagesToMem).
		OnEndFn(appendResultMessageToMem).
		Build()
	return compose.WithCallbacks(handler).DesignateNode(nodeKey)
}

func GetToolMemCallBackOption(nodeKey string) compose.Option {
	handler := callbacks.NewHandlerBuilder().
		OnEndFn(appendResultMessagesToMem).
		Build()
	return compose.WithCallbacks(handler).DesignateNode(nodeKey)
}

func setInputMessagesToMem(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	session := ctx.Value(SessionContextKey).(*Session)
	messages := input.([]*schema.Message)
	session.SetMessages(messages)
	return ctx
}

func appendResultMessageToMem(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	session := ctx.Value(SessionContextKey).(*Session)
	message := output.(*schema.Message)
	session.AppendMessages(message)
	return ctx
}

func appendResultMessagesToMem(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	session := ctx.Value(SessionContextKey).(*Session)
	messages := output.([]*schema.Message)
	session.AppendMessages(messages...)
	return ctx
}
