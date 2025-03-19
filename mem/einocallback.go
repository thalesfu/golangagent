package mem

import (
	"context"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
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
	session.InitNewAndSave("")
	ctx = context.WithValue(ctx, SessionContextKey, session)
	return ctx
}

func GetSessionFromContext(ctx context.Context) *Session {
	return ctx.Value(SessionContextKey).(*Session)
}

func GetModelMemCallBackOptionWithNodeKey(nodeKey string) compose.Option {
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(setInputMessagesToMem).
		OnEndFn(appendResultMessageToMem).
		Build()
	return compose.WithCallbacks(handler).DesignateNode(nodeKey)
}

func GetModelMemCallBackOptionWithNodePath(path ...*compose.NodePath) compose.Option {
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(setInputMessagesToMem).
		OnEndFn(appendResultMessageToMem).
		Build()
	return compose.WithCallbacks(handler).DesignateNodeWithPath().DesignateNodeWithPath(path...)
}

func setInputMessagesToMem(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	session := ctx.Value(SessionContextKey).(*Session)

	if in, ok := input.(*model.CallbackInput); ok {
		session.SetMessages(in.Messages)
	}

	return ctx
}

func appendResultMessageToMem(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	session := ctx.Value(SessionContextKey).(*Session)

	if out, ok := output.(*model.CallbackOutput); ok {
		session.AppendMessages(out.Message)
	}

	return ctx
}
