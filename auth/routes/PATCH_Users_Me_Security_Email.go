package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"dsoob/backend/tools"
)

func PATCH_Users_Me_Security_Email(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}

	var Body struct {
		Email string `json:"email" validate:"required,email"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}

	// Duplicate Check
	var UsageEmail int
	err := tools.Database.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM user WHERE email_address = LOWER(?)",
		Body.Email,
	).Scan(
		&UsageEmail,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if UsageEmail > 0 {
		tools.SendClientError(w, r, tools.ERROR_SIGNUP_DUPLICATE_EMAIL)
		return
	}

	// Update User
	var (
		UserEmailAddressPrevious  string
		UserEmailVerifyToken      = tools.GenerateTokenString()
		UserEmailVerifyExpiration = time.Now().Add(tools.TOKEN_LIFETIME_EMAIL_VERIFY)
	)

	tx, err := tools.Database.BeginTx(r.Context(), nil)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		"SELECT email_address FROM user WHERE id = ?",
		session.UserID,
	).Scan(
		&UserEmailAddressPrevious,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`UPDATE user SET
			updated			 	= CURRENT_TIMESTAMP,
			email_verified 		= FALSE,
			email_address 	 	= LOWER(?),
			token_verify 	 	= ?,
			token_verify_eat 	= ?
		WHERE id = ?`,
		Body.Email,
		UserEmailVerifyToken,
		UserEmailVerifyExpiration,
		session.UserID,
		session.UserID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Notify User
	go tools.EmailVerify(
		Body.Email,
		tools.LocalsEmailVerify{
			Token: UserEmailVerifyToken,
		},
	)
	go tools.EmailNotifyUserEmailModified(
		UserEmailAddressPrevious,
		tools.LocalsNotifyUserEmailModified{},
	)

	w.WriteHeader(http.StatusNoContent)
}
