package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"dsoob/backend/tools"
)

func PATCH_Users_Me_Security_Password(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		OldPassword string `json:"old_password" validate:"required,password"`
		NewPassword string `json:"new_password" validate:"required,password"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}

	session := tools.GetSession(r)
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Account Password Fields
	var (
		UserEmailAddress       string
		UserPasswordHash       *string
		UserPasswordHistoryRAW string
	)
	err := tools.Database.
		QueryRowContext(ctx, `SELECT email_address, password_hash, password_history FROM user WHERE id = $1`, session.UserID).
		Scan(&UserEmailAddress, &UserPasswordHash, &UserPasswordHistoryRAW)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Check Old Password
	if UserPasswordHash == nil {
		tools.SendClientError(w, r, tools.ERROR_LOGIN_PASSWORD_RESET)
		return
	}
	if ok, err := tools.ComparePasswordHash(*UserPasswordHash, Body.OldPassword); err != nil {
		tools.SendServerError(w, r, err)
		return
	} else if !ok {
		tools.SendClientError(w, r, tools.ERROR_MFA_PASSWORD_INCORRECT)
		return
	}

	// Update Password History
	UserPasswordHistory := strings.Split(UserPasswordHistoryRAW, tools.ARRAY_DELIMITER)
	for _, oldPassword := range UserPasswordHistory {
		if ok, err := tools.ComparePasswordHash(oldPassword, Body.NewPassword); err != nil {
			tools.SendServerError(w, r, err)
			return
		} else if !ok {
			tools.SendClientError(w, r, tools.ERROR_LOGIN_PASSWORD_ALREADY_USED)
			return
		}
	}
	newPasswordHash, err := tools.GeneratePasswordHash(Body.NewPassword)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	UserPasswordHistory = append(UserPasswordHistory, newPasswordHash)
	if len(UserPasswordHistory) > tools.PASSWORD_HISTORY_LIMIT {
		UserPasswordHistory = UserPasswordHistory[1:]
	}

	// Update User
	tag, err := tools.Database.ExecContext(ctx,
		`UPDATE user SET
			updated			 = CURRENT_TIMESTAMP,
			password_hash	 = $1,
			password_history = $2
		WHERE id = $3`,
		newPasswordHash,
		strings.Join(UserPasswordHistory, tools.ARRAY_DELIMITER),
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
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}

	// Notify User
	go tools.EmailNotifyUserPasswordModified(
		UserEmailAddress,
		tools.LocalsNotifyUserPasswordModified{},
	)

	w.WriteHeader(http.StatusNoContent)
}
