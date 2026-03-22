package common

import (
	"context"
	"log/slog"
)

// ContextHandler é um slog.Handler que injeta automaticamente campos de
// rastreabilidade (request_id, user_id) do context.Context em cada registro de log.
// Isso garante que todos os logs emitidos por use cases via slog.ErrorContext /
// slog.InfoContext incluam o contexto de rastreio sem precisar passar o logger
// explicitamente por toda a pilha de chamadas.
type ContextHandler struct {
	inner slog.Handler
}

func NewContextHandler(inner slog.Handler) *ContextHandler {
	return &ContextHandler{inner: inner}
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if rid, _ := ctx.Value(ContextKeyRequestID).(string); rid != "" {
		r.AddAttrs(slog.String("request_id", rid))
	}
	if uid, _ := ctx.Value(ContextKeyUserID).(string); uid != "" {
		r.AddAttrs(slog.String("user_id", uid))
	}
	return h.inner.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{inner: h.inner.WithGroup(name)}
}
