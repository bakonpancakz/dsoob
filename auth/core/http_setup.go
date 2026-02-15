package core

import (
	"dsoob/backend/routes"
	"dsoob/backend/tools"
	"net/http"
	"time"
)

func SetupMux() *http.ServeMux {

	var (
		mux       = http.NewServeMux()
		limitFILE = tools.NewBodyLimit(8 * 1024 * 1024) // 8MB
		limitJSON = tools.NewBodyLimit(16 * 1024)       // 16KB
		limitBLOB = tools.NewBodyLimit(32 * 1024)       // 32KB
		rateAuth  = tools.NewRatelimit(&tools.RatelimitOptions{
			Period: 5 * time.Minute,
			Limit:  25,
		})
		ratePublic = tools.NewRatelimit(&tools.RatelimitOptions{
			Period: 1 * time.Minute,
			Limit:  1000,
		})
		ratePrivate = tools.NewRatelimit(&tools.RatelimitOptions{
			Period: 5 * time.Minute,
			Limit:  100,
		})
		rateImages = tools.NewRatelimit(&tools.RatelimitOptions{
			Period: 5 * time.Minute,
			Limit:  10,
		})
	)

	// Auth
	mux.Handle("/auth/login", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Login, rateAuth, limitJSON),
	})
	mux.Handle("/auth/signup", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Signup, rateAuth, limitJSON),
	})
	mux.Handle("/auth/logout", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Logout, rateAuth, limitJSON, tools.UseSession),
	})
	mux.Handle("/auth/password-reset", tools.MethodHandler{
		http.MethodPost:  tools.Chain(routes.POST_Auth_ResetPassword, rateAuth, limitJSON),
		http.MethodPatch: tools.Chain(routes.PATCH_Auth_ResetPassword, rateAuth, limitJSON),
	})
	mux.Handle("/auth/verify-login", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_VerifyLogin, rateAuth, limitJSON),
	})
	mux.Handle("/auth/verify-email", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_VerifyEmail, rateAuth, limitJSON),
	})

	// User
	mux.Handle("/users/@me", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me, ratePrivate, tools.UseSession),
		http.MethodPatch:  tools.Chain(routes.PATCH_Users_Me, ratePrivate, limitJSON, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/avatar", tools.MethodHandler{
		http.MethodPut:    tools.Chain(routes.PUT_Users_Me_Avatar, rateImages, limitFILE, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Avatar, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/banner", tools.MethodHandler{
		http.MethodPut:    tools.Chain(routes.PUT_Users_Me_Banner, rateImages, limitFILE, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Banner, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/security/sessions", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me_Security_Sessions, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/security/sessions/{id}", tools.MethodHandler{
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_Sessions_ID, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/security/mfa/setup", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me_Security_MFA_Setup, ratePrivate, tools.UseSession),
		http.MethodPost:   tools.Chain(routes.POST_Users_Me_Security_MFA_Setup, ratePrivate, limitJSON, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_MFA_Setup, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/security/mfa/codes", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me_Security_MFA_Codes, ratePrivate, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_MFA_Codes, ratePrivate, tools.UseSession),
	})
	mux.Handle("/users/@me/security/escalate", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Users_Me_Security_Escalate, rateAuth, limitJSON, tools.UseSession),
	})
	mux.Handle("/users/@me/security/password", tools.MethodHandler{
		http.MethodPatch: tools.Chain(routes.PATCH_Users_Me_Security_Password, ratePrivate, limitJSON, tools.UseSession),
	})
	mux.Handle("/users/@me/security/email", tools.MethodHandler{
		http.MethodPost:  tools.Chain(routes.POST_Users_Me_Security_Email, ratePrivate, tools.UseSession),
		http.MethodPatch: tools.Chain(routes.PATCH_Users_Me_Security_Email, ratePrivate, limitJSON, tools.UseSession),
	})
	mux.Handle("/users/@me/settings", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me_Settings, ratePrivate, tools.UseSession),
		http.MethodPut: tools.Chain(routes.PUT_Users_Me_Settings, ratePrivate, limitBLOB, tools.UseSession),
	})

	// Public
	mux.Handle("/users/bulk", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Users_Bulk, ratePublic),
	})

	// Default 404 Handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tools.SendClientError(w, r, tools.ERROR_GENERIC_NOT_FOUND)
	})

	return mux
}
