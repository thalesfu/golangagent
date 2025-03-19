package golangagent

import (
	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"time"
)

func GetLangfuseHandler() (callbacks.Handler, func()) {
	return langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "http://localhost:3000",
		PublicKey: "pk-lf-781f7d44-d585-42b4-a10d-25433157e84e",
		SecretKey: "sk-lf-56e4d12c-da97-4623-81ff-e504c08426e8",
		Timeout:   20 * time.Millisecond,
	})
}
