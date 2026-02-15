package routes

import (
	"net/http"
	"strings"

	"dsoob/backend/tools"
)

func DELETE_Users_Me_Security_MFA_Codes(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Regenerate Recovery Codes
	recoveryCodes := tools.GenerateRecoveryCodes()
	tag, err := tools.Database.ExecContext(ctx,
		`UPDATE dsoob.users SET
			updated 		= CURRENT_TIMESTAMP,
			mfa_codes 		= $1,
			mfa_codes_used 	= 0
		WHERE id = $2 AND mfa_enabled = TRUE`,
		strings.Join(recoveryCodes, tools.ARRAY_DELIMITER),
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

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
		"recovery_codes": recoveryCodes,
	})
}
