package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"dsoob/backend/tools"
)

func DELETE_Users_Me_Avatar(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Update Account
	var AvatarHash *string
	err := tools.Database.
		QueryRowContext(ctx, "UPDATE user SET avatar_hash = NULL WHERE id = $1 RETURNING avatar_hash", session.UserID).
		Scan(&AvatarHash)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Delete Image
	if AvatarHash == nil {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_IMAGE)
		return
	}
	go func() {
		paths := tools.ImagePaths(tools.ImageOptionsAvatars, session.UserID, *AvatarHash)
		if err := tools.StoragePublicDelete(paths...); err != nil {
			tools.LoggerStorage.Data(tools.ERROR, "Failed to Delete Profile Avatar", map[string]any{
				"paths": paths,
				"error": err.Error(),
			})
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}
