package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"dsoob/backend/tools"
)

func PATCH_Users_Me(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Displayname      *string `json:"displayname" validate:"omitempty,displayname"`
		Subtitle         *string `json:"subtitle" validate:"omitempty,displayname"`
		Biography        *string `json:"biography" validate:"omitempty,description"`
		AccentBanner     *int    `json:"accent_banner" validate:"omitempty,color"`
		AccentBorder     *int    `json:"accent_border" validate:"omitempty,color"`
		AccentBackground *int    `json:"accent_background" validate:"omitempty,color"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}
	session := tools.GetSession(r)

	// Fetch User
	var (
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
			username,
			displayname,
			subtitle,
			biography,
			avatar_hash,
			banner_hash,
			accent_banner,
			accent_border,
			accent_background
		FROM dsoob.profiles
		WHERE id = $1`,
		session.UserID,
	).Scan(
		&UserName,
		&UserDisplayname,
		&UserSubtitle,
		&UserBiography,
		&UserAvatarHash,
		&UserBannerHash,
		&UserAccentBanner,
		&UserAccentBorder,
		&UserAccentBackground,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	} else if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Apply Edits
	edited := false
	if Body.Displayname != nil {
		if len(*Body.Displayname) == 0 {
			UserDisplayname = UserName
		} else {
			UserDisplayname = *Body.Displayname
		}
		edited = true
	}
	if Body.Subtitle != nil {
		if len(*Body.Subtitle) == 0 {
			UserSubtitle = nil
		} else {
			UserSubtitle = Body.Subtitle
		}
		edited = true
	}
	if Body.Biography != nil {
		if len(*Body.Biography) == 0 {
			UserBiography = nil
		} else {
			UserBiography = Body.Biography
		}
		edited = true
	}
	if Body.AccentBanner != nil {
		if *Body.AccentBanner == 0 {
			UserAccentBanner = nil
		} else {
			UserAccentBanner = Body.AccentBanner
		}
		edited = true
	}
	if Body.AccentBorder != nil {
		if *Body.AccentBorder == 0 {
			UserAccentBorder = nil
		} else {
			UserAccentBorder = Body.AccentBorder
		}
		edited = true
	}
	if Body.AccentBackground != nil {
		if *Body.AccentBackground == 0 {
			UserAccentBackground = nil
		} else {
			UserAccentBackground = Body.AccentBackground
		}
		edited = true
	}
	if !edited {
		tools.SendClientError(w, r, tools.ERROR_BODY_EMPTY)
		return
	}

	// Update User
	tag, err := tools.Database.ExecContext(r.Context(),
		`UPDATE dsoob.profiles SET
			updated 		  = CURRENT_TIMESTAMP,
			displayname 	  = $1,
			subtitle 		  = $2,
			biography		  = $3,
			accent_banner 	  = $4,
			accent_border	  = $5,
			accent_background = $6
		WHERE id = $7`,
		UserDisplayname,
		UserSubtitle,
		UserBiography,
		UserAccentBanner,
		UserAccentBorder,
		UserAccentBackground,
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

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
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
