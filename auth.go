package main

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func (app *App) requireSuperuserAuth(re *core.RequestEvent) error {
	authHeader := re.Request.Header.Get("Authorization")
	if authHeader == "" {
		return re.UnauthorizedError("Authorization header required", nil)
	}

	// Extract token from "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return re.UnauthorizedError("Invalid authorization format", nil)
	}

	token := parts[1]

	// Verify the token belongs to a superuser
	superuser, err := app.pb.FindAuthRecordByToken(token, "_superusers")
	if err != nil {
		return re.UnauthorizedError("Invalid or expired token", err)
	}

	if superuser == nil {
		return re.UnauthorizedError("Superuser authentication required", nil)
	}

	return nil
}
