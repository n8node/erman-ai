package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type workspaceOIDCTestUsers struct {
	user *model.User
}

func (f workspaceOIDCTestUsers) GetByID(context.Context, string) (*model.User, error) {
	return f.user, nil
}

func TestWorkspaceOIDCAuthorizationCodeFlow(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	verifiedAt := time.Now()
	cfg := &config.Config{
		Environment:                "development",
		WorkspaceOIDCClientID:      "erman-affine",
		WorkspaceOIDCClientSecret:  "test-client-secret",
		WorkspaceOIDCPrivateKeyB64: base64.StdEncoding.EncodeToString(der),
	}
	h := NewWorkspaceOIDCHandler(workspaceOIDCTestUsers{user: &model.User{
		ID:              "user-1",
		Email:           "user@example.com",
		Locale:          "ru",
		EmailVerifiedAt: &verifiedAt,
	}}, cfg)

	verifier := "workspace-pkce-verifier"
	sum := sha256.Sum256([]byte(verifier))
	authorizeQuery := url.Values{
		"client_id":             {"erman-affine"},
		"redirect_uri":          {cfg.WorkspaceOIDCCallbackURL()},
		"response_type":         {"code"},
		"scope":                 {"openid email profile"},
		"state":                 {"state-1"},
		"nonce":                 {"nonce-1"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
	}
	req := httptest.NewRequest(http.MethodGet, "/authorize?"+authorizeQuery.Encode(), nil)
	authMW := middleware.NewAuth("test-jwt-secret")
	session, err := authMW.IssueToken("user-1", "user", time.Hour)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: session})
	rec := httptest.NewRecorder()
	authMW.Required(http.HandlerFunc(h.Authorize)).ServeHTTP(rec, req)
	require.Equal(t, http.StatusFound, rec.Code)

	callback, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "state-1", callback.Query().Get("state"))
	code := callback.Query().Get("code")
	require.NotEmpty(t, code)

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {cfg.WorkspaceOIDCCallbackURL()},
		"code_verifier": {verifier},
	}
	tokenReq := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.SetBasicAuth("erman-affine", "test-client-secret")
	tokenRec := httptest.NewRecorder()
	h.Token(tokenRec, tokenReq)
	require.Equal(t, http.StatusOK, tokenRec.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(tokenRec.Body.Bytes(), &payload))
	idToken, ok := payload["id_token"].(string)
	require.True(t, ok)
	parsed, err := jwt.Parse(idToken, func(token *jwt.Token) (any, error) {
		return &privateKey.PublicKey, nil
	}, jwt.WithIssuer(cfg.WorkspaceOIDCIssuer()), jwt.WithAudience("erman-affine"))
	require.NoError(t, err)
	assert.True(t, parsed.Valid)
	claims := parsed.Claims.(jwt.MapClaims)
	assert.Equal(t, "user-1", claims["sub"])
	assert.Equal(t, "user@example.com", claims["email"])
	assert.Equal(t, "nonce-1", claims["nonce"])
}

func TestWorkspaceOIDCValidatesPlainPKCE(t *testing.T) {
	code := workspaceAuthorizationCode{
		codeChallenge: "expected",
		codeMethod:    "plain",
	}
	assert.True(t, validWorkspacePKCE(code, "expected"))
	assert.False(t, validWorkspacePKCE(code, "different"))
}
