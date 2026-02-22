package routes

import (
	"dsoob/backend/tools"
	"net/http"
)

func DELETE_Users_Me_Banner(w http.ResponseWriter, r *http.Request) {
	tools.EndpointImageDelete(w, r, "banner_hash", tools.ImageOptionsBanners)
}
