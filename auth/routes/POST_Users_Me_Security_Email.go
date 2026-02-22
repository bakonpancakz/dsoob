package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"dsoob/backend/tools"
)

func POST_Users_Me_Security_Email(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)

	// Update User
	var (
		UserEmailAddress          string
		UserEmailVerifyToken      = tools.GenerateTokenString()
		UserEmailVerifyExpiration = time.Now().Add(tools.TOKEN_LIFETIME_EMAIL_VERIFY)
	)
	err := tools.Database.QueryRowContext(r.Context(),
		`UPDATE user SET
			updated 		 = CURRENT_TIMESTAMP,
			token_verify 	 = ?,
			token_verify_eat = ?
		WHERE id = ? AND email_verified = FALSE
		RETURNING email_address`,
		UserEmailVerifyToken,
		UserEmailVerifyExpiration,
		session.UserID,
	).Scan(
		&UserEmailAddress,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_MFA_EMAIL_ALREADY_VERIFIED)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Notify User
	go tools.EmailVerify(
		UserEmailAddress,
		tools.LocalsEmailVerify{
			Token: UserEmailVerifyToken,
		},
	)

	w.WriteHeader(http.StatusNoContent)
}
