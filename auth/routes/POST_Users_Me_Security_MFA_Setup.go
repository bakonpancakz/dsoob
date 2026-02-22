package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func POST_Users_Me_Security_MFA_Setup(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Passcode string `json:"passcode" validate:"required,passcode"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}
	session := tools.GetSession(r)

	// Fetch User
	var (
		UserMFAEnabled bool
		UserMFASecret  *string
	)
	if err := tools.Database.QueryRowContext(r.Context(),
		"SELECT mfa_enabled, mfa_secret FROM user WHERE id = ?",
		session.UserID,
	).Scan(
		&UserMFAEnabled,
		&UserMFASecret,
	); err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if UserMFAEnabled {
		tools.SendClientError(w, r, tools.ERROR_MFA_SETUP_ALREADY)
		return
	}
	if UserMFASecret == nil {
		tools.SendClientError(w, r, tools.ERROR_MFA_SETUP_NOT_INITIALIZED)
		return
	}
	if !tools.ValidateTOTPCode(Body.Passcode, *UserMFASecret) {
		tools.SendClientError(w, r, tools.ERROR_MFA_PASSCODE_INCORRECT)
		return
	}

	// Update User
	if _, err := tools.Database.ExecContext(r.Context(),
		"UPDATE user SET mfa_enabled = TRUE WHERE id = ?",
		session.UserID,
	); err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
