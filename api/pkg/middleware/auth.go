package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
)

type AuthMiddleware struct {
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (a *AuthMiddleware) Auth(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")

	if authHeader == "" {
		return apperror.UnauthorizedError(errors.New("request without authorization header"), "Authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return apperror.UnauthorizedError(errors.New("invalid authorization header"), "Authorization header is invalid")
	}

	token := authHeader[7:]

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return apperror.UnauthorizedError(err, "Failed to create request to Google OAuth")
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return apperror.UnauthorizedError(err, "Failed to get profile from Google OAuth")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apperror.UnauthorizedError(errors.New("non-200 response from Google OAuth"), "Failed to get profile from Google OAuth")
	}

	var profile domain.Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return apperror.UnauthorizedError(err, "Failed to decode profile from Google OAuth")
	}

	ctx.Locals("profile", profile)
	return ctx.Next()
}

func (a *AuthMiddleware) Owner(ctx *fiber.Ctx) error {
	profile, ok := ctx.Locals("profile").(domain.Profile)
	if !ok {
		return apperror.UnauthorizedError(errors.New("profile not found in context"), "User profile is required")
	}

	if profile.Email != "6530162621@student.chula.ac.th" {
		return apperror.ForbiddenError(errors.New("who are you?"), "who are you?")
	}

	return ctx.Next()
}
