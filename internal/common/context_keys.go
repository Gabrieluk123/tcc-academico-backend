package common

// contextKey é um tipo privado para chaves de contexto neste pacote.
// O uso de um tipo privado evita colisão com chaves de outros pacotes.
type contextKey string

const (
	ContextKeyRequestID contextKey = "request_id"
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyRoleID    contextKey = "role_id"
)
