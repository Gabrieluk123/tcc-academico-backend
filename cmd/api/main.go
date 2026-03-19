package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

func main() {
	// =========================================================================
	// 1. Configuração do Logger (slog em formato JSON)
	// =========================================================================
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// =========================================================================
	// 2. Inicialização do Framework Web (Echo v5)
	// =========================================================================
	e := echo.New()

	// Rota básica de Health Check (para monitoramento de infraestrutura)
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "online",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// =========================================================================
	// 3. Injeção de Dependências (Wire-up)
	// =========================================================================
	// TODO: Inicializar a conexão com o PostgreSQL
	// TODO: Instanciar Repositórios (internal/infrastructure/database)
	// TODO: Instanciar HashProvider e TokenGenerator
	// TODO: Instanciar UseCases (internal/usecase/auth)
	// TODO: Instanciar Handlers (internal/http/handler)
	// TODO: Configurar Rotas (internal/http.SetupRoutes(e, ...))

	// =========================================================================
	// 4. Configuração do Servidor HTTP
	// =========================================================================
	srv := &http.Server{
		Addr:    ":8080", // Porta padrão da nossa API
		Handler: e,
	}

	// Iniciamos o servidor em uma Goroutine para ele não travar o fluxo principal
	go func() {
		slog.Info("iniciando servidor", slog.String("porta", "8080"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("erro critico no servidor", slog.String("erro", err.Error()))
			os.Exit(1)
		}
	}()

	// =========================================================================
	// 5. Graceful Shutdown
	// =========================================================================
	// Criamos um canal que escuta sinais de interrupção do Sistema Operacional (Ctrl+C ou Docker Stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// A aplicação fica travada aqui até receber um sinal
	<-quit
	slog.Info("sinal de interrupção recebido, iniciando desligamento...")

	// Damos um prazo máximo de 10 segundos para as requisições ativas terminarem
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("erro ao forçar desligamento", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	slog.Info("servidor finalizado com segurança")
}
