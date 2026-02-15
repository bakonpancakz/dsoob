package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"dsoob/backend/tools"
)

func DELETE_Users_Me_Banner(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Update Account
	var BannerHash *string
	err := tools.Database.
		QueryRowContext(ctx, "UPDATE user SET banner_hash = NULL WHERE id = $1 RETURNING banner_hash", session.UserID).
		Scan(&BannerHash)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Delete Image
	if BannerHash == nil {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_IMAGE)
		return
	}
	go func() {
		paths := tools.ImagePaths(tools.ImageOptionsBanners, session.UserID, *BannerHash)
		if err := tools.StoragePublicDelete(paths...); err != nil {
			tools.LoggerStorage.Data(tools.ERROR, "Failed to Delete Profile Banner", map[string]any{
				"paths": paths,
				"error": err.Error(),
			})
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}
