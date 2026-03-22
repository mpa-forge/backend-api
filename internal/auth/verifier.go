package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const defaultJWKSCacheTTL = 5 * time.Minute

// ClerkVerifier validates Clerk-issued session tokens against the instance JWKS.
type ClerkVerifier struct {
	issuer   string
	audience string
	client   *http.Client
	cacheTTL time.Duration

	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	keysExpire time.Time
}

// NewClerkVerifier builds a verifier that fetches signing keys from the Clerk
// frontend API JWKS endpoint derived from the configured issuer URL.
func NewClerkVerifier(issuerURL *url.URL, audience string) *ClerkVerifier {
	return &ClerkVerifier{
		issuer:   strings.TrimRight(issuerURL.String(), "/"),
		audience: audience,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cacheTTL: defaultJWKSCacheTTL,
	}
}

// VerifyToken validates the token signature and standard claims, then maps the
// verified Clerk claims into the internal principal used by handlers.
func (v *ClerkVerifier) VerifyToken(ctx context.Context, token string) (Principal, error) {
	claims := &sessionClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(parsed *jwt.Token) (any, error) {
		if parsed.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("%w: unsupported signing algorithm %q", ErrUnauthenticated, parsed.Method.Alg())
		}

		kid, _ := parsed.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("%w: missing key id", ErrUnauthenticated)
		}

		key, keyErr := v.publicKey(ctx, kid)
		if keyErr != nil {
			return nil, keyErr
		}

		return key, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience))
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %v", ErrUnauthenticated, err)
	}

	principal, err := claims.principal()
	if err != nil {
		return Principal{}, err
	}

	return principal, nil
}

func (v *ClerkVerifier) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	if key := v.cachedKey(keyID); key != nil {
		return key, nil
	}

	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}

	if key := v.cachedKey(keyID); key != nil {
		return key, nil
	}

	return nil, fmt.Errorf("%w: key %q not found in jwks", ErrUnauthenticated, keyID)
}

func (v *ClerkVerifier) cachedKey(keyID string) *rsa.PublicKey {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if time.Now().After(v.keysExpire) {
		return nil
	}

	if v.keys == nil {
		return nil
	}

	return v.keys[keyID]
}

func (v *ClerkVerifier) refreshKeys(ctx context.Context) error {
	keys, err := fetchJWKS(ctx, v.client, v.issuer)
	if err != nil {
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.keys = keys
	v.keysExpire = time.Now().Add(v.cacheTTL)

	return nil
}

type sessionClaims struct {
	jwt.RegisteredClaims
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	Role        string   `json:"role"`
	GivenName   string   `json:"given_name"`
	FamilyName  string   `json:"family_name"`
}

func (c *sessionClaims) principal() (Principal, error) {
	if c.Subject == "" {
		return Principal{}, fmt.Errorf("%w: missing subject claim", ErrUnauthenticated)
	}

	role, err := mapRole(c.Role, c.Roles)
	if err != nil {
		return Principal{}, err
	}

	return Principal{
		UserID:      c.Subject,
		Email:       c.Email,
		DisplayName: displayName(c.DisplayName, c.GivenName, c.FamilyName, c.Email, c.Subject),
		Role:        role,
	}, nil
}

func mapRole(role string, roles []string) (Role, error) {
	for _, candidate := range append([]string{role}, roles...) {
		switch Role(strings.ToLower(strings.TrimSpace(candidate))) {
		case "":
			continue
		case RoleAdmin:
			return RoleAdmin, nil
		case RoleUser:
			return RoleUser, nil
		default:
			return "", fmt.Errorf("%w: unsupported role %q", ErrForbidden, candidate)
		}
	}

	return RoleUser, nil
}

func displayName(explicit, given, family, email, subject string) string {
	if trimmed := strings.TrimSpace(explicit); trimmed != "" {
		return trimmed
	}

	fullName := strings.TrimSpace(strings.TrimSpace(given) + " " + strings.TrimSpace(family))
	if fullName != "" {
		return fullName
	}

	if localPart, _, found := strings.Cut(strings.TrimSpace(email), "@"); found && localPart != "" {
		return localPart
	}

	return subject
}
