package tools

import (
	"database/sql"
	"errors"
	"io"
	"math"
	"net/http"
)

func EndpointImageUpload(w http.ResponseWriter, r *http.Request, column string, options ImageOptions) {

	session := GetSession(r)
	if err := r.ParseMultipartForm(math.MaxInt64); err != nil {
		SendClientError(w, r, ERROR_BODY_INVALID_TYPE)
		return
	}

	// Collect incoming Data
	var (
		UploadData    []byte
		UploadHash    string
		UploadSuccess bool
		PreviousHash  *string
	)
	if f, _, err := r.FormFile("image"); err != nil {
		SendClientError(w, r, ERROR_BODY_INVALID_FIELD)
		return
	} else if data, err := io.ReadAll(f); err != nil {
		SendClientError(w, r, ERROR_BODY_INVALID_DATA)
		return
	} else {
		f.Close()
		UploadData = data
	}

	// Process Image
	if ok, hash := ImageHandler(w, r, options, session.UserID, UploadData); !ok {
		return
	} else {
		UploadHash = hash
	}
	defer func() {

		// Delete leftover images from a failed upload
		if !UploadSuccess && UploadHash != "" {
			paths := ImagePaths(options, session.UserID, UploadHash)
			if err := StoragePublicDelete(paths...); err != nil {
				LoggerStorage.Data(ERROR, "Unable to delete leftover images", map[string]any{
					"paths": paths,
					"error": err.Error(),
				})
			}
		}

		// Delete previous images (if any)
		if UploadSuccess && PreviousHash != nil && *PreviousHash != UploadHash {
			paths := ImagePaths(options, session.UserID, *PreviousHash)
			if err := StoragePublicDelete(paths...); err != nil {
				LoggerStorage.Data(ERROR, "Failed to delete previous images", map[string]any{
					"paths": paths,
					"error": err.Error(),
				})
			}
		}
	}()

	// Update User
	tx, err := Database.BeginTx(r.Context(), nil)
	if err != nil {
		SendServerError(w, r, err)
		return
	}

	err = tx.QueryRowContext(r.Context(),
		"SELECT "+column+" FROM user WHERE id = ?",
		session.UserID,
	).Scan(
		&PreviousHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		SendClientError(w, r, ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		SendServerError(w, r, err)
		return
	}

	_, err = tx.ExecContext(r.Context(),
		"UPDATE user SET updated = CURRENT_TIMESTAMP, "+column+" = ? WHERE id = ?",
		UploadHash,
		session.UserID,
	)
	if err != nil {
		SendServerError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		SendServerError(w, r, err)
		return
	}

	UploadSuccess = true
	w.WriteHeader(http.StatusNoContent)
}

func EndpointImageDelete(w http.ResponseWriter, r *http.Request, column string, options ImageOptions) {

	var BannerHash *string
	session := GetSession(r)

	// Update Account
	tx, err := Database.BeginTx(r.Context(), nil)
	if err != nil {
		SendServerError(w, r, err)
		return
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(r.Context(),
		"SELECT "+column+" FROM user WHERE id = ?",
		session.UserID,
	).Scan(
		&BannerHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		SendClientError(w, r, ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		SendServerError(w, r, err)
		return
	}

	_, err = tx.ExecContext(r.Context(),
		"UPDATE user SET updated = CURRENT_TIMESTAMP, "+column+" = NULL WHERE id = ?",
		session.UserID,
	)
	if err != nil {
		SendServerError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		SendServerError(w, r, err)
		return
	}

	// Delete Previous Image (if any)
	if BannerHash == nil {
		SendClientError(w, r, ERROR_UNKNOWN_IMAGE)
		return
	}
	go func() {
		paths := ImagePaths(options, session.UserID, *BannerHash)
		if err := StoragePublicDelete(paths...); err != nil {
			LoggerStorage.Data(ERROR, "Failed to delete images", map[string]any{
				"paths": paths,
				"error": err.Error(),
			})
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}
