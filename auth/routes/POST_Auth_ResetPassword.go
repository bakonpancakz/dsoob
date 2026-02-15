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
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Update User
	var (
		ResetTokenExpiration = time.Now().Add(tools.LIFETIME_TOKEN_EMAIL_RESET)
		ResetToken           = tools.GenerateSignedString()
		UserID               int64
		UserEmailAddress     string
	)
	err := tools.Database.QueryRowContext(ctx,
		`UPDATE user SET
			updated 		= CURRENT_TIMESTAMP,
			token_reset_eat = $1,
			token_reset 	= $2
		WHERE email_address = LOWER($3)
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
