package main

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func (app *App) requireSuperuserAuth(re *core.RequestEvent) error {
	app.pb.Logger().Info("=== requireSuperuserAuth called ===")

	authHeader := re.Request.Header.Get("Authorization")
	app.pb.Logger().Info("Auth header", "header", authHeader)

	if authHeader == "" {
		app.pb.Logger().Error("No authorization header")
		return re.UnauthorizedError("Authorization header required", nil)
	}

	// Extract token from "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	app.pb.Logger().Info("Auth header parts", "count", len(parts), "first", parts[0])

	if len(parts) != 2 || parts[0] != "Bearer" {
		app.pb.Logger().Error("Invalid authorization format", "parts", parts)
		return re.UnauthorizedError("Invalid authorization format", nil)
	}

	token := parts[1]
	app.pb.Logger().Info("Extracted token", "token_length", len(token), "token_start", token[:20])

	// Verify the token belongs to a superuser
	app.pb.Logger().Info("About to call FindAuthRecordByToken")
	superuser, err := app.pb.FindAuthRecordByToken(token, "_superusers")

	if err != nil {
		app.pb.Logger().Error("FindAuthRecordByToken failed", "error", err)
		return re.UnauthorizedError("Invalid or expired token", err)
	}

	if superuser == nil {
		app.pb.Logger().Error("Superuser is nil")
		return re.UnauthorizedError("Superuser authentication required", nil)
	}

	app.pb.Logger().Info("Authentication successful", "superuser_id", superuser.Id)
	return nil
}
