package routes

import (
	"dsoob/backend/tools"
	"net/http"
)

func DELETE_Users_Me_Avatar(w http.ResponseWriter, r *http.Request) {
	tools.EndpointImageDelete(w, r, "avatar_hash", tools.ImageOptionsAvatars)
}
