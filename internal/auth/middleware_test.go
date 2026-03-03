package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func newMiddlewareTestApp(secret string) *fiber.App {
	app := fiber.New()
	app.Get("/protected", JWTMiddleware(secret), func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"email":   c.Locals("email"),
		})
	})
	return app
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	app := newMiddlewareTestApp(secret)

	token, err := generateToken("user::abc", "user@example.com", secret)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	app := newMiddlewareTestApp(secret)

	claims := Claims{
		UserID: "user::abc",
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)

	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_MalformedToken(t *testing.T) {
	app := newMiddlewareTestApp("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")

	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	app := newMiddlewareTestApp("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_WrongScheme(t *testing.T) {
	secret := "test-secret"
	app := newMiddlewareTestApp(secret)

	token, _ := generateToken("user::abc", "user@example.com", secret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+token)

	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
