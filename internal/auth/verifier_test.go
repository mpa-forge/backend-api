package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestClerkVerifierVerifyToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	var hits atomic.Int64
	issuerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/.well-known/jwks.json" {
			http.NotFound(w, r)
			return
		}

		_ = json.NewEncoder(w).Encode(jwksDocument{
			Keys: []jsonWebKey{rsaJWK("test-key", &privateKey.PublicKey)},
		})
	}))
	defer issuerServer.Close()

	issuerURL, err := url.Parse(issuerServer.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	verifier := NewClerkVerifier(issuerURL, "https://api.local.mpa-forge")
	verifier.cacheTTL = time.Hour

	validToken := signedToken(t, privateKey, map[string]any{
		"kid": "test-key",
	}, map[string]any{
		"iss":          issuerServer.URL,
		"sub":          "user_123",
		"aud":          []string{"https://api.local.mpa-forge"},
		"exp":          time.Now().Add(time.Hour).Unix(),
		"nbf":          time.Now().Add(-time.Minute).Unix(),
		"iat":          time.Now().Unix(),
		"email":        "user@example.com",
		"display_name": "Example User",
		"role":         "admin",
	})

	principal, err := verifier.VerifyToken(context.Background(), validToken)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
	if principal.UserID != "user_123" || principal.Role != RoleAdmin {
		t.Fatalf("principal = %#v, want user_123/admin", principal)
	}

	_, err = verifier.VerifyToken(context.Background(), validToken)
	if err != nil {
		t.Fatalf("VerifyToken() second call error = %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("jwks hits = %d, want 1 due to cache", hits.Load())
	}

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name: "wrong audience",
			token: signedToken(t, privateKey, map[string]any{"kid": "test-key"}, map[string]any{
				"iss": issuerServer.URL,
				"sub": "user_123",
				"aud": []string{"https://other.example.com"},
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: ErrUnauthenticated,
		},
		{
			name: "wrong issuer",
			token: signedToken(t, privateKey, map[string]any{"kid": "test-key"}, map[string]any{
				"iss": "https://wrong-issuer.example.com",
				"sub": "user_123",
				"aud": []string{"https://api.local.mpa-forge"},
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: ErrUnauthenticated,
		},
		{
			name: "unsupported role",
			token: signedToken(t, privateKey, map[string]any{"kid": "test-key"}, map[string]any{
				"iss":  issuerServer.URL,
				"sub":  "user_123",
				"aud":  []string{"https://api.local.mpa-forge"},
				"exp":  time.Now().Add(time.Hour).Unix(),
				"role": "viewer",
			}),
			wantErr: ErrForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := verifier.VerifyToken(context.Background(), test.token)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("VerifyToken() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func signedToken(t *testing.T, privateKey *rsa.PrivateKey, header, claims map[string]any) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(claims))
	for key, value := range header {
		token.Header[key] = value
	}

	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return signed
}

func rsaJWK(keyID string, publicKey *rsa.PublicKey) jsonWebKey {
	return jsonWebKey{
		KeyType: "RSA",
		KeyID:   keyID,
		Use:     "sig",
		N:       base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		E:       base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
	}
}
