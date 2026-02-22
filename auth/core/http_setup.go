package core

import (
	"dsoob/backend/routes"
	"dsoob/backend/tools"
	"net/http"
	"time"
)

func SetupMux() *http.ServeMux {

	var (
		mux                 = http.NewServeMux()
		limitFILE           = tools.NewBodyLimit(8 * 1024 * 1024)   // 8MB
		limitJSON           = tools.NewBodyLimit(16 * 1024)         // 16KB
		limitBLOB           = tools.NewBodyLimit(32 * 1024)         // 32KB
		rateAuthSignup      = tools.NewRatelimit(3, 24*time.Hour)   // Limit: New Accounts
		rateAuthLogin       = tools.NewRatelimit(5, 5*time.Minute)  // Limit: Login Attempts
		rateAuthVerify      = tools.NewRatelimit(5, 5*time.Minute)  // Limit: Escalation / Password Reset Attempts
		ratePublicRead      = tools.NewRatelimit(50, 1*time.Minute) // Limit: Public Requests
		ratePrivateRead     = tools.NewRatelimit(50, 5*time.Minute) // Limit: User Read Requests
		ratePrivateWrite    = tools.NewRatelimit(10, 5*time.Minute) // Limit: User Write Requests
		ratePrivateSpammy   = tools.NewRatelimit(5, 30*time.Minute) // Limit: Requests that should not be spammed
		rateImagesReadWrite = tools.NewRatelimit(10, 5*time.Minute) // Limit: User Images
	)

	// Auth
	mux.Handle("/auth/login", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Login, rateAuthLogin, limitJSON),
	})
	mux.Handle("/auth/signup", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Signup, rateAuthSignup, limitJSON),
	})
	mux.Handle("/auth/logout", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Logout, rateAuthLogin, limitJSON, tools.UseSession),
	})
	mux.Handle("/auth/password-reset", tools.MethodHandler{
		http.MethodPost:  tools.Chain(routes.POST_Auth_ResetPassword, rateAuthVerify, limitJSON),
		http.MethodPatch: tools.Chain(routes.PATCH_Auth_ResetPassword, rateAuthVerify, limitJSON),
	})
	mux.Handle("/auth/verify-login", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_VerifyLogin, rateAuthVerify, limitJSON),
	})
	mux.Handle("/auth/verify-email", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_VerifyEmail, rateAuthVerify, limitJSON),
	})

	// User
	mux.Handle("/users/@me", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me, ratePrivateRead, tools.UseSession),
		http.MethodPatch:  tools.Chain(routes.PATCH_Users_Me, ratePrivateWrite, limitJSON, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me, ratePrivateWrite, tools.UseSession),
	})
	mux.Handle("/users/@me/avatar", tools.MethodHandler{
		http.MethodPut:    tools.Chain(routes.PUT_Users_Me_Avatar, rateImagesReadWrite, limitFILE, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Avatar, rateImagesReadWrite, tools.UseSession),
	})
	mux.Handle("/users/@me/banner", tools.MethodHandler{
		http.MethodPut:    tools.Chain(routes.PUT_Users_Me_Banner, rateImagesReadWrite, limitFILE, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Banner, rateImagesReadWrite, tools.UseSession),
	})
	mux.Handle("/users/@me/security/sessions", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me_Security_Sessions, ratePrivateRead, tools.UseSession),
	})
	mux.Handle("/users/@me/security/sessions/{id}", tools.MethodHandler{
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_Sessions_ID, ratePrivateWrite, tools.UseSession),
	})
	mux.Handle("/users/@me/security/mfa/setup", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me_Security_MFA_Setup, ratePrivateRead, tools.UseSession),
		http.MethodPost:   tools.Chain(routes.POST_Users_Me_Security_MFA_Setup, ratePrivateWrite, limitJSON, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_MFA_Setup, ratePrivateWrite, tools.UseSession),
	})
	mux.Handle("/users/@me/security/mfa/codes", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me_Security_MFA_Codes, ratePrivateRead, tools.UseSession),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_MFA_Codes, ratePrivateWrite, tools.UseSession),
	})
	mux.Handle("/users/@me/security/escalate", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Users_Me_Security_Escalate, rateAuthVerify, limitJSON, tools.UseSession),
	})
	mux.Handle("/users/@me/security/password", tools.MethodHandler{
		http.MethodPatch: tools.Chain(routes.PATCH_Users_Me_Security_Password, ratePrivateWrite, limitJSON, tools.UseSession),
	})
	mux.Handle("/users/@me/security/email", tools.MethodHandler{
		http.MethodPost:  tools.Chain(routes.POST_Users_Me_Security_Email, ratePrivateSpammy, tools.UseSession),
		http.MethodPatch: tools.Chain(routes.PATCH_Users_Me_Security_Email, ratePrivateSpammy, limitJSON, tools.UseSession),
	})
	mux.Handle("/users/@me/settings", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me_Settings, ratePrivateRead, tools.UseSession),
		http.MethodPut: tools.Chain(routes.PUT_Users_Me_Settings, ratePrivateWrite, limitBLOB, tools.UseSession),
	})

	// Public
	mux.Handle("/users/bulk", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Users_Bulk, ratePublicRead),
	})
	mux.Handle("/users/{id}/keychain", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_ID_Keychain, ratePublicRead),
	})

	// Default 404 Handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tools.SendClientError(w, r, tools.ERROR_GENERIC_NOT_FOUND)
	})

	return mux
}
