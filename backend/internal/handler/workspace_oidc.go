package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

const workspaceOIDCKeyID = "erman-workspace-1"

type workspaceOIDCUserRepository interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
}

type workspaceAuthorizationCode struct {
	user          *model.User
	redirectURI   string
	nonce         string
	codeChallenge string
	codeMethod    string
	expiresAt     time.Time
}

type WorkspaceOIDCHandler struct {
	users        workspaceOIDCUserRepository
	issuer       string
	clientID     string
	clientSecret string
	redirectURI  string
	privateKey   *rsa.PrivateKey

	mu    sync.Mutex
	codes map[string]workspaceAuthorizationCode
}

func NewWorkspaceOIDCHandler(
	users workspaceOIDCUserRepository,
	cfg *config.Config,
) *WorkspaceOIDCHandler {
	return &WorkspaceOIDCHandler{
		users:        users,
		issuer:       cfg.WorkspaceOIDCIssuer(),
		clientID:     cfg.WorkspaceOIDCClientID,
		clientSecret: cfg.WorkspaceOIDCClientSecret,
		redirectURI:  cfg.WorkspaceOIDCCallbackURL(),
		privateKey:   parseWorkspaceOIDCPrivateKey(cfg.WorkspaceOIDCPrivateKeyB64),
		codes:        make(map[string]workspaceAuthorizationCode),
	}
}

func (h *WorkspaceOIDCHandler) Discovery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"issuer":                                h.issuer,
		"authorization_endpoint":                h.issuer + "/authorize",
		"token_endpoint":                        h.issuer + "/token",
		"userinfo_endpoint":                     h.issuer + "/userinfo",
		"jwks_uri":                              h.issuer + "/jwks",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "email", "profile"},
		"claims_supported":                      []string{"sub", "email", "email_verified", "name", "preferred_username", "locale"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
		"code_challenge_methods_supported":      []string{"S256", "plain"},
	})
}

func (h *WorkspaceOIDCHandler) JWKS(w http.ResponseWriter, _ *http.Request) {
	if h.privateKey == nil {
		writeError(w, http.StatusServiceUnavailable, "workspace OIDC is not configured")
		return
	}
	exponentBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(exponentBytes, uint32(h.privateKey.PublicKey.E))
	exponentBytes = []byte(strings.TrimLeft(string(exponentBytes), "\x00"))
	writeJSON(w, http.StatusOK, map[string]any{
		"keys": []map[string]string{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": workspaceOIDCKeyID,
			"n":   base64.RawURLEncoding.EncodeToString(h.privateKey.PublicKey.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(exponentBytes),
		}},
	})
}

func (h *WorkspaceOIDCHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeError(w, http.StatusServiceUnavailable, "workspace OIDC is not configured")
		return
	}
	query := r.URL.Query()
	if query.Get("client_id") != h.clientID ||
		query.Get("redirect_uri") != h.redirectURI ||
		query.Get("response_type") != "code" ||
		!hasScope(query.Get("scope"), "openid") {
		writeError(w, http.StatusBadRequest, "invalid authorization request")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user.IsBlocked || !user.EmailVerified() {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	code, err := randomWorkspaceToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create authorization code")
		return
	}
	h.mu.Lock()
	h.deleteExpiredCodes(time.Now())
	h.codes[code] = workspaceAuthorizationCode{
		user:          user,
		redirectURI:   h.redirectURI,
		nonce:         query.Get("nonce"),
		codeChallenge: query.Get("code_challenge"),
		codeMethod:    query.Get("code_challenge_method"),
		expiresAt:     time.Now().Add(90 * time.Second),
	}
	h.mu.Unlock()

	callback, _ := url.Parse(h.redirectURI)
	params := callback.Query()
	params.Set("code", code)
	if state := query.Get("state"); state != "" {
		params.Set("state", state)
	}
	callback.RawQuery = params.Encode()
	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func (h *WorkspaceOIDCHandler) Token(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeError(w, http.StatusServiceUnavailable, "workspace OIDC is not configured")
		return
	}
	if err := r.ParseForm(); err != nil {
		writeOIDCError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	clientID, clientSecret := r.Form.Get("client_id"), r.Form.Get("client_secret")
	if basicID, basicSecret, ok := r.BasicAuth(); ok {
		clientID, clientSecret = basicID, basicSecret
	}
	if clientID != h.clientID || clientSecret != h.clientSecret {
		writeOIDCError(w, http.StatusUnauthorized, "invalid_client")
		return
	}
	if r.Form.Get("grant_type") != "authorization_code" {
		writeOIDCError(w, http.StatusBadRequest, "unsupported_grant_type")
		return
	}

	codeValue := r.Form.Get("code")
	h.mu.Lock()
	code, ok := h.codes[codeValue]
	if ok {
		delete(h.codes, codeValue)
	}
	h.mu.Unlock()
	if !ok || time.Now().After(code.expiresAt) ||
		r.Form.Get("redirect_uri") != code.redirectURI ||
		!validWorkspacePKCE(code, r.Form.Get("code_verifier")) {
		writeOIDCError(w, http.StatusBadRequest, "invalid_grant")
		return
	}

	now := time.Now()
	expiresAt := now.Add(15 * time.Minute)
	accessClaims := jwt.MapClaims{
		"iss": h.issuer, "sub": code.user.ID, "aud": h.clientID,
		"iat": now.Unix(), "exp": expiresAt.Unix(), "token_use": "access",
		"email": code.user.Email, "email_verified": true,
		"name": code.user.Email, "preferred_username": code.user.Email, "locale": code.user.Locale,
	}
	idClaims := jwt.MapClaims{
		"iss": h.issuer, "sub": code.user.ID, "aud": h.clientID,
		"iat": now.Unix(), "exp": expiresAt.Unix(),
		"email": code.user.Email, "email_verified": true,
		"name": code.user.Email, "preferred_username": code.user.Email, "locale": code.user.Locale,
	}
	if code.nonce != "" {
		idClaims["nonce"] = code.nonce
	}
	accessToken, err := h.sign(accessClaims)
	if err != nil {
		writeOIDCError(w, http.StatusInternalServerError, "server_error")
		return
	}
	idToken, err := h.sign(idClaims)
	if err != nil {
		writeOIDCError(w, http.StatusInternalServerError, "server_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(time.Until(expiresAt).Seconds()),
		"id_token":     idToken,
		"scope":        "openid email profile",
	})
}

func (h *WorkspaceOIDCHandler) UserInfo(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if raw == "" || raw == r.Header.Get("Authorization") {
		writeOIDCError(w, http.StatusUnauthorized, "invalid_token")
		return
	}
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() || h.privateKey == nil {
			return nil, errors.New("unexpected signing method")
		}
		return &h.privateKey.PublicKey, nil
	}, jwt.WithIssuer(h.issuer), jwt.WithAudience(h.clientID))
	if err != nil || !token.Valid {
		writeOIDCError(w, http.StatusUnauthorized, "invalid_token")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["token_use"] != "access" {
		writeOIDCError(w, http.StatusUnauthorized, "invalid_token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sub":                claims["sub"],
		"email":              claims["email"],
		"email_verified":     claims["email_verified"],
		"name":               claims["name"],
		"preferred_username": claims["preferred_username"],
		"locale":             claims["locale"],
	})
}

func (h *WorkspaceOIDCHandler) configured() bool {
	return h.privateKey != nil && h.clientID != "" && h.clientSecret != ""
}

func (h *WorkspaceOIDCHandler) sign(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = workspaceOIDCKeyID
	return token.SignedString(h.privateKey)
}

func (h *WorkspaceOIDCHandler) deleteExpiredCodes(now time.Time) {
	for code, value := range h.codes {
		if now.After(value.expiresAt) {
			delete(h.codes, code)
		}
	}
}

func parseWorkspaceOIDCPrivateKey(encoded string) *rsa.PrivateKey {
	if encoded == "" {
		return nil
	}
	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		der, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil {
		return nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey
		}
	}
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key
	}
	return nil
}

func validWorkspacePKCE(code workspaceAuthorizationCode, verifier string) bool {
	if code.codeChallenge == "" {
		return true
	}
	switch code.codeMethod {
	case "", "plain":
		return verifier == code.codeChallenge
	case "S256":
		sum := sha256.Sum256([]byte(verifier))
		return base64.RawURLEncoding.EncodeToString(sum[:]) == code.codeChallenge
	default:
		return false
	}
}

func hasScope(scopes, wanted string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == wanted {
			return true
		}
	}
	return false
}

func randomWorkspaceToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func writeOIDCError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
