package routes

import (
	"crypto/sha256"
	"dsoob/backend/tools"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
)

func GET_Users_Me_Settings(w http.ResponseWriter, r *http.Request) {
	session := tools.GetSession(r)

	// Fetch File
	filepath := path.Join(tools.DATA_DIRECTORY, "settings", strconv.FormatInt(session.UserID, 10)+".raw")
	raw, err := os.ReadFile(filepath)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Check Client Cache
	storedHash := fmt.Sprintf("%x", sha256.Sum256(raw))
	clientHash := r.Header.Get("If-None-Match")
	if storedHash == clientHash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Return Results
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(raw)))
	w.Header().Set("ETag", storedHash)
	w.Write(raw)
}
