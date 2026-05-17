package static

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestVerifier_Validate(t *testing.T) {
	v := New("secret-token", "u", nil)
	_, err := v.Validate(context.Background(), "wrong")
	if err != auth.ErrUnauthorized {
		t.Fatalf("want unauthorized, got %v", err)
	}
	sub, err := v.Validate(context.Background(), "secret-token")
	if err != nil || sub.Sub != "u" {
		t.Fatalf("sub=%+v err=%v", sub, err)
	}
}
