package auth

import "testing"

func TestEnforcer_RBAC(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		RBACEnabled: true,
		RoleReader:  "veil-reader",
		RoleAdmin:   "veil-admin",
	}
	e := NewEnforcer(cfg)

	reader := &Subject{Roles: []string{"veil-reader"}}
	if err := e.Enforce(reader, PermGraphRead); err != nil {
		t.Fatalf("reader: %v", err)
	}

	none := &Subject{Roles: []string{"other"}}
	if err := e.Enforce(none, PermGraphRead); err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}

	off := NewEnforcer(Config{Enabled: true, RBACEnabled: false})
	if err := off.Enforce(none, PermGraphRead); err != nil {
		t.Fatalf("rbac off: %v", err)
	}
}
