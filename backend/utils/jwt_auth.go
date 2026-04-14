package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// JWKS cache
var (
	jwksCache     *jose.JSONWebKeySet
	jwksMutex     sync.RWMutex
	jwksLastFetch time.Time
)

const (
	projectID       = "ffaea088-baeb-4606-a299-7d0ee52226f0"
	jwksRefreshTime = 6 * time.Hour
)

type JWTClaims struct {
	jwt.Claims
}
type TokenExpiredError struct {
	ExpiredAt time.Time
	Message   string
}

// Error implements the error interface
func (e *TokenExpiredError) Error() string {
	return fmt.Sprintf("token expired at %v: %s", e.ExpiredAt, e.Message)
}

// fetchJWKS fetches the JWKS from Stack Auth
func fetchJWKS() (*jose.JSONWebKeySet, error) {
	jwksMutex.RLock()
	if jwksCache != nil && time.Since(jwksLastFetch) < jwksRefreshTime {
		defer jwksMutex.RUnlock()
		return jwksCache, nil
	}
	jwksMutex.RUnlock()

	// Need to fetch or refresh JWKS
	jwksMutex.Lock()
	defer jwksMutex.Unlock()

	if jwksCache != nil && time.Since(jwksLastFetch) < jwksRefreshTime {
		return jwksCache, nil
	}

	jwksURL := fmt.Sprintf("https://api.stack-auth.com/api/v1/projects/%s/.well-known/jwks.json", projectID)
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status code %d", resp.StatusCode)
	}

	keySet := new(jose.JSONWebKeySet)
	err = json.NewDecoder(resp.Body).Decode(keySet)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	jwksCache = keySet
	jwksLastFetch = time.Now()
	return keySet, nil
}

func ExtractToken(r *http.Request) (string, error) {
	apiKeyHeader := r.Header.Get("X-API-Key")
	if apiKeyHeader == "" {
		return "", errors.New("authorization api key is missing")
	}

	return apiKeyHeader, nil
}

func JWTExtractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header format must be bearer {token}")
	}
	return parts[1], nil

}

func getKeyID(token *jwt.JSONWebToken) (string, bool) {
	if token == nil {
		return "", false
	}

	if len(token.Headers) == 0 {
		return "", false
	}

	kid := token.Headers[0].KeyID
	return kid, kid != ""
}

// validates the Authorization API Key
func ValidateAPIKey(apiKey string) (*JWTClaims, error) {
	var id string
	var expireAt sql.NullTime

	err := DB.QueryRow(`
        SELECT id, expire_at
        FROM api_keys
        WHERE key_hash = $1 AND is_active = true 
    `, apiKey).Scan(&id, &expireAt)

	if err != nil {
		return nil, errors.New("not a valid authorization api key")
	}

	// If expire_at is set and in the past, the key is expired
	if expireAt.Valid && time.Now().After(expireAt.Time) {
		return nil, fmt.Errorf("API Key expired at %v", expireAt.Time)
	}

	// Return id and expiry (can be nil if not set)
	claims := &JWTClaims{}
	if expireAt.Valid {

		claims.Subject = id
		claims.Issuer = "Trugen"
		var exp *jwt.NumericDate
		if expireAt.Valid {
			exp = jwt.NewNumericDate(expireAt.Time)
		}
		claims.Expiry = exp
	}
	return claims, nil
}

// validates the JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	// Basic format validation (JWT should have 3 parts separated by dots)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format: not a valid JWT")
	}

	// Fetch JWKS
	jwks, err := fetchJWKS()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Parse the JWT with allowed signature algorithms
	tok, err := jwt.ParseSigned(tokenString, []jose.SignatureAlgorithm{
		jose.ES256, jose.RS256, jose.RS384, jose.RS512,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Find the key used to sign the JWT
	kid, found := getKeyID(tok)
	if !found {
		return nil, errors.New("JWT missing key ID")
	}

	var key jose.JSONWebKey
	keyFound := false
	for _, k := range jwks.Keys {
		if k.KeyID == kid {
			key = k
			keyFound = true
			break
		}
	}

	if !keyFound || key.Key == nil {
		return nil, errors.New("unable to find key with matching kid")
	}

	// Verify the JWT and extract claims
	claims := &JWTClaims{}
	if err := tok.Claims(key.Key, claims); err != nil {
		return nil, fmt.Errorf("failed to verify JWT: %w", err)
	}

	// Validate expiration with special handling for expired tokens
	if err := claims.Validate(jwt.Expected{
		Time: time.Now(),
	}); err != nil {
		// Check specifically for expired token error
		if strings.Contains(err.Error(), "token is expired") {
			return nil, &TokenExpiredError{
				ExpiredAt: claims.Expiry.Time(),
				Message:   "token has expired, please use the refresh token endpoint",
			}
		}
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	return claims, nil
}

// JWTAuthMiddleware is a middleware that validates JWT tokens
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := JWTExtractToken(r)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate token
		claims, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddleware validates the API Key and returns a JSON error if unauthorized
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tokenString, err := ExtractToken(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Bad Request: " + err.Error(),
			})
			return
		}

		claims, err := ValidateAPIKey(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized: " + err.Error(),
			})
			return
		}

		ctx := context.WithValue(r.Context(), "apiKeyId", claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
