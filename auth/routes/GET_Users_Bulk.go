package routes

import (
	"dsoob/backend/tools"
	"net/http"
	"time"
)

func POST_Users_Bulk(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		UserIDs []int64 `json:"user_ids" validate:"min=1,max=100"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Users
	rows, err := tools.Database.QueryContext(ctx, `
		SELECT
			u.id,
			u.created,
			u.username,
			u.displayname,
			u.subtitle,
			u.biography,
			u.avatar_hash,
			u.banner_hash,
			u.accent_banner,
			u.accent_border,
			u.accent_background,
			COALESCE(ARRAY_AGG(s.device_public_key) FILTER (WHERE s.user_id IS NOT NULL), '{}') AS keys
		FROM user u
		LEFT JOIN user_session s ON s.user_id = u.id
		WHERE u.id = ANY($1)
		GROUP BY u.id;`,
		Body.UserIDs,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	defer rows.Close()

	var (
		UserItems            = make([]map[string]any, 0, 100)
		UserID               int64
		UserCreated          time.Time
		UserName             string
		UserDisplayname      string
		UserSubtitle         *string
		UserBiography        *string
		UserAvatarHash       *string
		UserBannerHash       *string
		UserAccentBanner     *int
		UserAccentBorder     *int
		UserAccentBackground *int
		UserPublicKeys       []string
	)
	for rows.Next() {
		if err := rows.Scan(
			&UserID,
			&UserCreated,
			&UserName,
			&UserDisplayname,
			&UserSubtitle,
			&UserBiography,
			&UserAvatarHash,
			&UserBannerHash,
			&UserAccentBanner,
			&UserAccentBorder,
			&UserAccentBackground,
			&UserPublicKeys,
		); err != nil {
			tools.SendServerError(w, r, err)
			return
		}

		UserItems = append(UserItems, map[string]any{
			"id":                UserID,
			"created":           UserCreated,
			"username":          UserName,
			"displayname":       UserDisplayname,
			"subtitle":          UserSubtitle,
			"biography":         UserBiography,
			"avatar":            UserAvatarHash,
			"banner":            UserBannerHash,
			"accent_banner":     UserAccentBanner,
			"accent_border":     UserAccentBorder,
			"accent_background": UserAccentBackground,
			"public_keys":       UserPublicKeys,
		})

	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, UserItems)
}
