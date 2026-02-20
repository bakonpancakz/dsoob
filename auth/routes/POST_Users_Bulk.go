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

	// Fetch Users
	rows, err := tools.Database.QueryContext(r.Context(),
		`SELECT
			id, created, username, displayname,
			subtitle, biography, avatar_hash, banner_hash,
			accent_banner, accent_border, accent_background
		FROM user WHERE id = IN($1)`,
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
		})

	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, UserItems)
}
