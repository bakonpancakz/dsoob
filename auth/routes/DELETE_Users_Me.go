package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"dsoob/backend/tools"
)

func DELETE_Users_Me(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}

	// Delete Relevant Account
	var (
		UserID           int64
		UserEmailAddress string
		UserAvatarHash   *string
		UserBannerHash   *string
	)
	err := tools.Database.QueryRowContext(r.Context(),
		"DELETE FROM user WHERE id = ? RETURNING id, email_address, avatar_hash, banner_hash",
		session.UserID,
	).Scan(
		&UserID,
		&UserEmailAddress,
		&UserAvatarHash,
		&UserBannerHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Delete Any Account Images
	var imagePaths = []string{}
	if UserAvatarHash != nil {
		imagePaths = append(imagePaths,
			tools.ImagePaths(tools.ImageOptionsAvatars, UserID, *UserAvatarHash)...,
		)
	}
	if UserBannerHash != nil {
		imagePaths = append(imagePaths,
			tools.ImagePaths(tools.ImageOptionsBanners, UserID, *UserBannerHash)...,
		)
	}
	go func() {
		if err := tools.StoragePublicDelete(imagePaths...); err != nil {
			tools.LoggerStorage.Data(tools.ERROR, "Failed to Delete Account Images", map[string]any{
				"paths": imagePaths,
				"error": err.Error(),
			})
		}
	}()

	// Notify User
	go tools.EmailNotifyUserDeleted(UserEmailAddress,
		tools.LocalsNotifyUserDeleted{
			Reason: "User Request",
		},
	)

	w.WriteHeader(http.StatusNoContent)
}
