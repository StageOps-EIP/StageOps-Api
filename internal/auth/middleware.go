package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware validates the Bearer token from the Authorization header.
// On success it sets "user_id" and "email" in fiber.Ctx locals.
func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return respondError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentification requise.")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return respondError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Format d'autorisation invalide.")
		}

		claims, err := validateToken(parts[1], secret)
		if err != nil {
			return respondError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Token invalide ou expiré.")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}
