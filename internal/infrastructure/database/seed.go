package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"academico/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// catalogPermissions defines the full permission catalog for the system.
// Format: "resource:action". Add new entries here as new modules are built.
var catalogPermissions = []PermissionDB{
	// Users module
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000001"), Slug: "user:create", Description: "Criar usuários"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000002"), Slug: "user:read", Description: "Visualizar usuários"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000003"), Slug: "user:update", Description: "Editar usuários"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000004"), Slug: "user:delete", Description: "Desativar usuários"},

	// Roles module
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000010"), Slug: "role:create", Description: "Criar perfis"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000011"), Slug: "role:read", Description: "Visualizar perfis"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000012"), Slug: "role:update", Description: "Editar perfis"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000013"), Slug: "role:delete", Description: "Excluir perfis"},

	// Permissions module (for the admin UI that assigns permissions to roles)
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000020"), Slug: "permission:read", Description: "Listar permissões disponíveis"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000021"), Slug: "permission:assign", Description: "Atribuir permissões a perfis"},

	// Sessions module
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000030"), Slug: "session:read", Description: "Visualizar sessões ativas"},
	{ID: uuid.MustParse("01960000-0000-7000-a000-000000000031"), Slug: "session:revoke", Description: "Revogar sessões de usuários"},
}

// SeedPermissions inserts the permission catalog into the database idempotently.
// Safe to call on every application boot.
func SeedPermissions(db *gorm.DB) error {
	if err := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}}, // Conflito baseado no campo "slug"
			DoNothing: true,
		}).
		Create(&catalogPermissions).Error; err != nil {
		return fmt.Errorf("falha ao seedar permissões: %w", err)
	}
	slog.Info("permissões seedadas com sucesso", slog.Int("total", len(catalogPermissions)))
	return nil
}

// SeedAdminRole ensures the "Admin" role exists and that the Casbin enforcer
// holds a wildcard policy giving it full access to every resource and action.
func SeedAdminRole(db *gorm.DB, enforcer domain.Enforcer) error {
	adminRole := RoleDB{
		ID:          uuid.MustParse("01960000-0000-7000-b000-000000000001"),
		Name:        "Admin",
		Description: "Administrador do sistema com acesso total",
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&adminRole).Error; err != nil {
		return fmt.Errorf("falha ao seedar role Admin: %w", err)
	}

	added, err := enforcer.AddPolicy(context.Background(), "Admin", "*", "*")
	if err != nil {
		return fmt.Errorf("falha ao adicionar policy do Admin no Casbin: %w", err)
	}
	if added {
		slog.Info("policy Admin:*:* adicionada ao Casbin")
	}

	return nil
}

// SeedSystemAdmin cria o usuário admin inicial de forma idempotente e segura.
// Usa variáveis de ambiente para email/senha, fallback para padrão.
func SeedSystemAdmin(ctx context.Context, userUseCase domain.UserUseCase) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		err := fmt.Errorf("variável de ambiente ADMIN_EMAIL não definida")
		return err
	}
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		err := fmt.Errorf("variável de ambiente ADMIN_PASSWORD não definida")
		return err
	}

	existing, err := userUseCase.FindUserByEmail(ctx, adminEmail)
	if err != nil {
		return fmt.Errorf("falha ao buscar usuário admin: %w", err)
	}
	if existing != nil {
		slog.InfoContext(ctx, "usuário admin já existe", slog.String("email", adminEmail))
		return nil
	}

	adminUser := &domain.User{
		ID:           uuid.MustParse("01960000-0000-7000-c000-000000000001"),
		FirstName:    "SYS",
		LastName:     "ADMIN",
		PasswordHash: adminPassword,
		RoleID:       uuid.MustParse("01960000-0000-7000-b000-000000000001"),
		Email:        adminEmail,
		IsActive:     true,
	}
	if err := userUseCase.CreateUser(ctx, adminUser); err != nil {
		return fmt.Errorf("falha ao criar usuário admin: %w", err)
	}
	slog.InfoContext(ctx, "usuário admin seedado com sucesso", slog.String("email", adminEmail))
	return nil
}
