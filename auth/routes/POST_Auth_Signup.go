package routes

import (
	"net/http"
	"time"

	"dsoob/backend/tools"
)

func POST_Auth_Signup(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Email    string `json:"email" validate:"required,email"`
		Username string `json:"username" validate:"required,username"`
		Password string `json:"password" validate:"required,password"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}

	// Check for Duplicate Email or Username
	var UsageUsername, UsageEmail int
	if err := tools.Database.QueryRowContext(r.Context(),
		`SELECT
			(SELECT COUNT(*) FROM user WHERE username      = LOWER(?)),
			(SELECT COUNT(*) FROM user WHERE email_address = LOWER(?))`,
		Body.Username,
		Body.Email,
	).Scan(
		&UsageUsername,
		&UsageEmail,
	); err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if UsageUsername > 0 {
		tools.SendClientError(w, r, tools.ERROR_SIGNUP_DUPLICATE_USERNAME)
		return
	}
	if UsageEmail > 0 {
		tools.SendClientError(w, r, tools.ERROR_SIGNUP_DUPLICATE_EMAIL)
		return
	}

	// Create User
	var (
		UserID                = tools.GenerateSnowflake()
		UserEmailVerifyToken  = tools.GenerateTokenString()
		UserPasswordHash, err = tools.GeneratePasswordHash(Body.Password)
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if _, err := tools.Database.ExecContext(r.Context(),
		`INSERT INTO user (
			id,
			email_address,
			ip_address,
			password_hash,
			password_history,
			token_verify,
			token_verify_eat,
			username,
			displayname
		) VALUES (?, LOWER(?), ?, ?, ?, ?, ?, LOWER(?), ?)`,
		UserID,
		Body.Email,
		tools.GetRemoteIP(r),
		UserPasswordHash,
		UserEmailVerifyToken,
		time.Now().Add(tools.TOKEN_LIFETIME_EMAIL_VERIFY),
		Body.Username,
		Body.Username,
	); err != nil {
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

	w.WriteHeader(http.StatusNoContent)
}
