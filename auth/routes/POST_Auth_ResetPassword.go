package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"dsoob/backend/tools"
)

func POST_Auth_ResetPassword(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Email string `json:"email" validate:"required,email"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}

	// Update User
	var (
		ResetTokenExpiration = time.Now().Add(tools.TOKEN_LIFETIME_EMAIL_RESET)
		ResetToken           = tools.GenerateTokenString()
		UserID               int64
		UserEmailAddress     string
	)
	err := tools.Database.QueryRowContext(r.Context(),
		`UPDATE user SET
			updated 		= CURRENT_TIMESTAMP,
			token_reset_eat = ?,
			token_reset 	= ?
		WHERE email_address = LOWER(?)
		RETURNING id, email_address`,
		ResetTokenExpiration,
		ResetToken,
		Body.Email,
	).Scan(
		&UserID,
		&UserEmailAddress,
	)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Notify User
	go tools.EmailLoginForgotPassword(
		UserEmailAddress,
		tools.LocalsLoginForgotPassword{
			Token: ResetToken,
		},
	)

	w.WriteHeader(http.StatusNoContent)
}
