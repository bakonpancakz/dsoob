package routes

import (
	"database/sql"
	"errors"
	"io"
	"math"
	"net/http"

	"dsoob/backend/tools"
)

func PUT_Users_Me_Banner(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if err := r.ParseMultipartForm(math.MaxInt64); err != nil {
		tools.SendClientError(w, r, tools.ERROR_BODY_INVALID_TYPE)
		return
	}

	// Store Incoming Binary
	var (
		UploadOptions = tools.ImageOptionsBanners
		UploadData    []byte
		UploadHash    string
		UploadSuccess bool
	)
	if f, _, err := r.FormFile("image"); err != nil {
		tools.SendClientError(w, r, tools.ERROR_BODY_INVALID_FIELD)
		return
	} else if data, err := io.ReadAll(f); err != nil {
		tools.SendClientError(w, r, tools.ERROR_BODY_INVALID_DATA)
		return
	} else {
		f.Close()
		UploadData = data
	}

	// Process Image
	if ok, hash := tools.ImageHandler(w, r, UploadOptions, session.UserID, UploadData); !ok {
		return
	} else {
		UploadHash = hash
	}
	defer func() {
		// Delete any possibly leftover files from a failed upload
		if !UploadSuccess && UploadHash != "" {
			paths := tools.ImagePaths(UploadOptions, session.UserID, UploadHash)
			if err := tools.StoragePublicDelete(paths...); err != nil {
				tools.LoggerStorage.Data(tools.ERROR, "Failed to delete leftover banners", map[string]any{
					"paths": paths,
					"error": err.Error(),
				})
			}
		}
	}()

	// Store updated image hash
	var PreviousHash *string
	err := tools.Database.QueryRowContext(r.Context(),
		`UPDATE dsoob.profiles SET
			updated 	= CURRENT_TIMESTAMP,
			banner_hash = $1
		WHERE id = $2
		RETURNING banner_hash`,
		PreviousHash,
		session.UserID,
	).Scan(
		&PreviousHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	UploadSuccess = true

	// Delete Previous User Images
	go func() {
		if PreviousHash != nil {
			paths := tools.ImagePaths(UploadOptions, session.UserID, *PreviousHash)
			if err := tools.StoragePublicDelete(paths...); err != nil {
				tools.LoggerStorage.Data(tools.ERROR, "Failed to delete previous banners", map[string]any{
					"paths": paths,
					"error": err.Error(),
				})
			}
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}
