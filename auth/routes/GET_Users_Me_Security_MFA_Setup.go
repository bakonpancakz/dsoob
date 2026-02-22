package routes

import (
	"fmt"
	"net/http"
	"strings"

	"dsoob/backend/tools"
)

func GET_Users_Me_Security_MFA_Setup(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}

	// Fetch User
	var (
		UserEmailAddress string
		UserMFAEnabled   bool
		UserName         string
	)
	err := tools.Database.QueryRowContext(r.Context(),
		"SELECT email_address, mfa_enabled, username FROM user WHERE id = ?",
		session.UserID,
	).Scan(
		&UserEmailAddress,
		&UserMFAEnabled,
		&UserName,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if UserMFAEnabled {
		tools.SendClientError(w, r, tools.ERROR_MFA_SETUP_ALREADY)
		return
	}

	// Update User
	setupCodes := tools.GenerateRecoveryCodes()
	setupSecret := tools.GenerateTOTPSecret()
	setupURI := tools.GenerateTOTPURI(
		tools.SITE_NAME,
		fmt.Sprintf("%s (%s)", UserName, UserEmailAddress),
		setupSecret,
	)

	tag, err := tools.Database.ExecContext(r.Context(),
		`UPDATE user SET
			updated 		= CURRENT_TIMESTAMP,
			mfa_enabled 	= false,
			mfa_secret 		= ?,
			mfa_codes 		= ?,
			mfa_codes_used 	= 0
		WHERE id = ?`,
		setupSecret,
		strings.Join(setupCodes, tools.ARRAY_DELIMITER),
		session.UserID,
	)
	if c, err := tag.RowsAffected(); err != nil {
		tools.SendServerError(w, r, err)
		return
	} else if c == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
		"recovery_codes": setupCodes,
		"secret":         setupSecret,
		"uri":            setupURI,
	})
}
