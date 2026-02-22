package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func POST_Auth_VerifyEmail(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Token string `json:"token" validate:"required,token"`
	}
	if !tools.BindJSON(w, r, &Body) {
		return
	}

	// Update User
	tag, err := tools.Database.ExecContext(r.Context(),
		`UPDATE user SET
			updated 		 = CURRENT_TIMESTAMP,
			email_verified   = TRUE,
			token_verify 	 = NULL,
			token_verify_eat = NULL
		WHERE token_verify = ? AND token_verify_eat > CURRENT_TIMESTAMP`,
		Body.Token,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if c, err := tag.RowsAffected(); err != nil {
		tools.SendServerError(w, r, err)
		return
	} else if c == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
