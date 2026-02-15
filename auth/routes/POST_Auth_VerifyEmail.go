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
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Update User
	tag, err := tools.Database.ExecContext(ctx,
		`UPDATE user SET
			updated 		 = CURRENT_TIMESTAMP,
			email_verified   = TRUE,
			token_verify 	 = NULL,
			token_verify_eat = NULL
		WHERE token_verify = $1 AND token_verify_eat > NOW()`,
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
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_TOKEN)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
