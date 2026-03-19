package auth

import (
	"context"
	"fmt"

	"academico/internal/domain"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// rbacModel defines a 3-field RBAC policy: sub (role), obj (resource), act (action).
// keyMatch supports wildcards: e.g. "diary:*" matches "diary:create".
const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch(r.obj, p.obj) && keyMatch(r.act, p.act)
`

type casbinEnforcer struct {
	e *casbin.Enforcer
}

var _ domain.Enforcer = (*casbinEnforcer)(nil)

// NewCasbinEnforcer creates a Casbin enforcer backed by the GORM adapter,
// persisting policies in the casbin_rule table of the application database.
func NewCasbinEnforcer(db *gorm.DB) (domain.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar adapter Casbin: %w", err)
	}

	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar model Casbin: %w", err)
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar Casbin enforcer: %w", err)
	}

	// Load policies from database into memory on startup.
	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("falha ao carregar policies do banco: %w", err)
	}

	return &casbinEnforcer{e: e}, nil
}

func (c *casbinEnforcer) Enforce(_ context.Context, roleName, resource, action string) (bool, error) {
	allowed, err := c.e.Enforce(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar autorização: %w", err)
	}
	return allowed, nil
}

func (c *casbinEnforcer) AddPolicy(_ context.Context, roleName, resource, action string) (bool, error) {
	added, err := c.e.AddPolicy(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("erro ao adicionar policy: %w", err)
	}
	return added, nil
}

func (c *casbinEnforcer) RemovePolicy(_ context.Context, roleName, resource, action string) (bool, error) {
	removed, err := c.e.RemovePolicy(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("erro ao remover policy: %w", err)
	}
	return removed, nil
}
