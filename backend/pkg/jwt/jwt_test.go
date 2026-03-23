package jwt

import (
	"os"
	"testing"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := GenerateToken(42, "worker")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	mapClaims, ok := claims.(gjwt.MapClaims)
	if !ok {
		t.Fatalf("claims type = %T, want jwt.MapClaims", claims)
	}
	if got := mapClaims["role"]; got != "worker" {
		t.Fatalf("role = %v, want worker", got)
	}
	if got := uint(mapClaims["user_id"].(float64)); got != 42 {
		t.Fatalf("user_id = %d, want 42", got)
	}
}

func TestGenerateTokenMissingSecret(t *testing.T) {
	_ = os.Unsetenv("JWT_SECRET")
	if _, err := GenerateToken(1, "worker"); err == nil {
		t.Fatal("GenerateToken() expected error when JWT_SECRET is missing")
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret-a")
	token, err := GenerateToken(7, "worker")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	t.Setenv("JWT_SECRET", "secret-b")

	if _, err := ValidateToken(token); err == nil {
		t.Fatal("ValidateToken() expected error for wrong secret")
	}
}

func TestValidateTokenRejectsNoneAlgorithm(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	unsigned := gjwt.NewWithClaims(gjwt.SigningMethodNone, gjwt.MapClaims{
		"user_id": 1,
		"role":    "worker",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	token, err := unsigned.SignedString(gjwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ValidateToken(token); err == nil {
		t.Fatal("ValidateToken() expected error for non-HMAC algorithm")
	}
}
