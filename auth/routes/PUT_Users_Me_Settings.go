package routes

import (
	"crypto/sha256"
	"dsoob/backend/tools"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

func PUT_Users_Me_Settings(w http.ResponseWriter, r *http.Request) {
	session := tools.GetSession(r)

	// Collect Settings
	givenSettings, err := io.ReadAll(r.Body)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	givenHash := fmt.Sprintf("%x", sha256.Sum256(givenSettings))

	// Store Settings
	filepath := path.Join(tools.DATA_DIRECTORY, "settings", strconv.FormatInt(session.UserID, 10)+".raw")
	if err := os.WriteFile(filepath, givenSettings, tools.FILEMODE_SECURE); err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Return Results
	w.Header().Set("ETag", givenHash)
	w.WriteHeader(http.StatusNoContent)
}
