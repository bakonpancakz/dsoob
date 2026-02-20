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
		"SELECT COUNT(*) FROM user WHERE email_address = LOWER($1)",
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
		UserEmailVerifyToken      = tools.GenerateSignedString()
		UserEmailVerifyExpiration = time.Now().Add(tools.LIFETIME_TOKEN_EMAIL_VERIFY)
	)
	err = tools.Database.
		QueryRowContext(r.Context(),
			`UPDATE user SET
				updated			 	= CURRENT_TIMESTAMP,
				email_verified 		= FALSE,
				email_address 	 	= LOWER($1),
				token_verify 	 	= $2,
				token_verify_eat 	= $3
			WHERE id = $4
			RETURNING (SELECT email_address FROM user WHERE id = $4)`,
			Body.Email,
			UserEmailVerifyToken,
			UserEmailVerifyExpiration,
			session.UserID,
		).
		Scan(&UserEmailAddressPrevious)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
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
