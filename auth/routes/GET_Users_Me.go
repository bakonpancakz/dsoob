package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"dsoob/backend/tools"
)

func GET_Users_Me(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)

	// Fetch User
	var (
		UserID               int64
		UserCreated          time.Time
		UserEmailAddress     string
		UserEmailVerified    bool
		UserMFAEnabled       bool
		UserName             string
		UserDisplayname      string
		UserSubtitle         *string
		UserBiography        *string
		UserAvatarHash       *string
		UserBannerHash       *string
		UserAccentBanner     *int
		UserAccentBorder     *int
		UserAccentBackground *int
	)
	err := tools.Database.QueryRowContext(r.Context(),
		`SELECT
			id, created, email_address, email_verified, mfa_enabled,
			username, displayname, subtitle, biography,
			avatar_hash, banner_hash,
			accent_banner, accent_border, accent_background
		FROM user WHERE id = ?`,
		session.UserID,
	).Scan(
		&UserID, &UserCreated, &UserEmailAddress, &UserEmailVerified, &UserMFAEnabled,
		&UserName, &UserDisplayname, &UserSubtitle, &UserBiography,
		&UserAvatarHash, &UserBannerHash,
		&UserAccentBanner, &UserAccentBorder, &UserAccentBackground,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
		"id":                UserID,
		"created":           UserCreated,
		"email_address":     UserEmailAddress,
		"email_verified":    UserEmailVerified,
		"mfa_enabled":       UserMFAEnabled,
		"username":          UserName,
		"displayname":       UserDisplayname,
		"subtitle":          UserSubtitle,
		"biography":         UserBiography,
		"avatar":            UserAvatarHash,
		"banner":            UserBannerHash,
		"accent_banner":     UserAccentBanner,
		"accent_border":     UserAccentBorder,
		"accent_background": UserAccentBackground,
	})
}
