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

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"

	routes "academico/internal/http"
	"academico/internal/http/handler"
	infraAuth "academico/internal/infrastructure/auth"
	"academico/internal/infrastructure/database"
	usecaseAuth "academico/internal/usecase/auth"
)

func main() {
	// =========================================================================
	// 1. Configuração do Logger
	// =========================================================================
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// =========================================================================
	// 2. Carregar variáveis de ambiente (.env)
	// =========================================================================
	if err := godotenv.Load(); err != nil {
		slog.Warn("arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080" // Fallback seguro
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET não está configurado no ambiente")
		os.Exit(1)
	}

	// =========================================================================
	// 3. Conexão com Banco de Dados e Migrations
	// =========================================================================
	db, err := database.NewPostgresConnection()
	if err != nil {
		slog.Error("falha fatal ao conectar no banco de dados", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	slog.Info("executando automigrate do banco de dados...")
	if err := db.AutoMigrate(&database.RoleDB{}, &database.UserDB{}, &database.RefreshTokenDB{}); err != nil {
		slog.Error("falha ao rodar migrations", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	// =========================================================================
	// 4. Injeção de Dependências (Wire-up)
	// =========================================================================

	// A. Inicializa a Infraestrutura Externa (Providers e Banco)
	hashProvider := infraAuth.NewHashProvider()
	tokenGenerator := infraAuth.NewJWTGenerator(jwtSecret, 24*time.Hour)

	roleRepo := database.NewRoleRepository(db)
	userRepo := database.NewUserRepository(db)
	refreshRepo := database.NewRefreshTokenRepository(db)

	// B. Inicializa as Regras de Negócio (Use Cases)
	roleUC := usecaseAuth.NewRoleUseCase(roleRepo)
	userUC := usecaseAuth.NewUserUseCase(userRepo, hashProvider)
	authUC := usecaseAuth.NewAuthUseCase(userRepo, tokenGenerator, hashProvider, refreshRepo)

	// C. Inicializa os Controladores da Web (Handlers)
	roleHandler := handler.NewRoleHandler(roleUC)
	userHandler := handler.NewUserHandler(userUC)
	authHandler := handler.NewAuthHandler(authUC)

	// =========================================================================
	// 5. Inicialização do Framework Web (Echo v5) e Rotas
	// =========================================================================
	e := echo.New()

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "online",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	routes.SetupRoutes(e, authHandler, userHandler, roleHandler)

	// =========================================================================
	// 6. Configuração do Servidor HTTP
	// =========================================================================
	srv := &http.Server{
		Addr:    ":" + apiPort, // Usa a porta do .env
		Handler: e,
	}

	go func() {
		slog.Info("iniciando servidor", slog.String("porta", apiPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("erro critico no servidor", slog.String("erro", err.Error()))
			os.Exit(1)
		}
	}()

	// =========================================================================
	// 7. Graceful Shutdown
	// =========================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("sinal de interrupção recebido, iniciando desligamento...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("erro ao forçar desligamento", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	slog.Info("servidor finalizado com segurança")
}
