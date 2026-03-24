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

	"academico/internal/common"
	routes "academico/internal/http"
	"academico/internal/http/handler"
	"academico/internal/http/middleware"
	infraAuth "academico/internal/infrastructure/auth"
	"academico/internal/infrastructure/database"
	usecaseAuth "academico/internal/usecase/auth"

	echomw "github.com/labstack/echo/v5/middleware"
)

func main() {
	// =========================================================================
	// 1. Configuração do Logger
	// =========================================================================
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(common.NewContextHandler(jsonHandler))
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
	if err := db.AutoMigrate(
		&database.RoleDB{},
		&database.UserDB{},
		&database.RefreshTokenDB{},
		&database.PermissionDB{},
	); err != nil {
		slog.Error("falha ao rodar migrations", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	if err := database.SeedPermissions(db); err != nil {
		slog.Error("falha ao seedar permissões", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	// =========================================================================
	// 4. Injeção de Dependências (Wire-up)
	// =========================================================================

	// A. Inicializa a Infraestrutura Externa (Providers e Banco)
	hashProvider := infraAuth.NewHashProvider()
	tokenGenerator := infraAuth.NewJWTGenerator(jwtSecret, 24*time.Hour)

	enforcer, err := infraAuth.NewCasbinEnforcer(db)
	if err != nil {
		slog.Error("falha ao inicializar Casbin", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	permissionRepo := database.NewPermissionRepository(db)
	roleRepo := database.NewRoleRepository(db, permissionRepo)
	userRepo := database.NewUserRepository(db, permissionRepo)
	refreshRepo := database.NewRefreshTokenRepository(db)

	if err := database.SeedAdminRole(db, enforcer); err != nil {
		slog.Error("falha ao seedar role Admin", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	// B. Inicializa as Regras de Negócio (Use Cases)
	roleUC := usecaseAuth.NewRoleUseCase(roleRepo, permissionRepo, userRepo)
	userUC := usecaseAuth.NewUserUseCase(userRepo, hashProvider, permissionRepo)
	authUC := usecaseAuth.NewAuthUseCase(userRepo, tokenGenerator, hashProvider, refreshRepo, 24*time.Hour)
	permissionUC := usecaseAuth.NewPermissionUseCase(permissionRepo)

	if err := database.SeedSystemAdmin(context.Background(), userUC); err != nil {
		slog.Error("falha ao seedar usuário admin", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	// C. Inicializa os Controladores da Web (Handlers)
	roleHandler := handler.NewRoleHandler(roleUC)
	userHandler := handler.NewUserHandler(userUC)
	authHandler := handler.NewAuthHandler(authUC)
	permissionHandler := handler.NewPermissionHandler(permissionUC)

	// =========================================================================
	// 5. Inicialização do Framework Web (Echo v5) e Rotas
	// =========================================================================
	e := echo.New()

	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{"Authorization", "Content-Type", "Accept", "X-Request-ID"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	// Recover captura panics em qualquer handler e retorna 500 em vez de
	// derrubar o processo — usa o middleware nativo do Echo v5.
	e.Use(echomw.Recover())
	e.Use(middleware.NewRequestIDMiddleware())

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "online",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	jwtMW := middleware.NewJWTMiddleware(jwtSecret, refreshRepo)
	routes.SetupRoutes(e, authHandler, userHandler, roleHandler, permissionHandler, jwtMW, enforcer, roleRepo)

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
