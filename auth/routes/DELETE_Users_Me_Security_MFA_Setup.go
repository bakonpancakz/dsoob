package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func DELETE_Users_Me_Security_MFA_Setup(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}

	// Reset Fields
	tag, err := tools.Database.ExecContext(r.Context(),
		`UPDATE user SET
			updated 		= CURRENT_TIMESTAMP,
			mfa_enabled 	= false,
			mfa_secret	 	= NULL,
			mfa_codes 		= '',
			mfa_codes_used 	= 0
		WHERE mfa_enabled = TRUE AND id = ?`,
		session.UserID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if c, err := tag.RowsAffected(); err != nil {
		tools.SendServerError(w, r, err)
		return
	} else if c == 0 {
		tools.SendClientError(w, r, tools.ERROR_MFA_DISABLED)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
