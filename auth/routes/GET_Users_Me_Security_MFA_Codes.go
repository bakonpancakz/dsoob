package routes

import (
	"net/http"
	"strings"

	"dsoob/backend/tools"
)

func GET_Users_Me_Security_MFA_Codes(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch User
	var (
		UserMFAEnabled  bool
		UserMFACodesRAW string
	)
	err := tools.Database.
		QueryRowContext(ctx, "SELECT mfa_enabled, mfa_codes FROM user WHERE id = $1", session.UserID).
		Scan(&UserMFAEnabled, &UserMFACodesRAW)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if !UserMFAEnabled {
		tools.SendClientError(w, r, tools.ERROR_MFA_DISABLED)
		return
	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
		"recovery_codes": strings.Split(UserMFACodesRAW, tools.ARRAY_DELIMITER),
	})
}
