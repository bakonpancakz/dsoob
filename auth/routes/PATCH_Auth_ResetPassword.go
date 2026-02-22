package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"dsoob/backend/tools"
)

func PATCH_Auth_ResetPassword(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		NewPassword string `json:"password" validate:"required,password"`
		Token       string `json:"token" validate:"required,token"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}

	// Fetch User
	var (
		UserID                 int64
		UserEmailAddress       string
		UserPasswordHistoryRAW string
	)
	err := tools.Database.QueryRowContext(r.Context(),
		`SELECT id, email_address, password_history
		FROM user WHERE token_reset = ? AND token_reset_eat > CURRENT_TIMESTAMP`,
		Body.Token,
	).Scan(
		&UserID,
		&UserEmailAddress,
		&UserPasswordHistoryRAW,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Update Password History
	UserPasswordHistory := strings.Split(UserPasswordHistoryRAW, tools.ARRAY_DELIMITER)
	for _, oldPassword := range UserPasswordHistory {
		if ok, err := tools.ComparePasswordHash(oldPassword, Body.NewPassword); err != nil {
			tools.SendServerError(w, r, err)
			return
		} else if ok {
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
	tag, err := tools.Database.ExecContext(r.Context(),
		`UPDATE user SET
			updated 		 = CURRENT_TIMESTAMP,
			token_reset 	 = NULL,
			token_reset_eat	 = NULL,
			password_hash 	 = ?,
			password_history = ?
		WHERE id = ?`,
		newPasswordHash,
		strings.Join(UserPasswordHistory, tools.ARRAY_DELIMITER),
		UserID,
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

	// Alert User
	go tools.EmailNotifyUserPasswordModified(
		UserEmailAddress,
		tools.LocalsNotifyUserPasswordModified{},
	)

	w.WriteHeader(http.StatusNoContent)
}
