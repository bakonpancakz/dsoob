package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func PUT_Users_Me_Banner(w http.ResponseWriter, r *http.Request) {
	tools.EndpointImageUpload(w, r, "banner_hash", tools.ImageOptionsBanners)
}
