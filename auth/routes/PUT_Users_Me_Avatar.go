package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func PUT_Users_Me_Avatar(w http.ResponseWriter, r *http.Request) {
	tools.EndpointImageUpload(w, r, "avatar_hash", tools.ImageOptionsAvatars)
}
