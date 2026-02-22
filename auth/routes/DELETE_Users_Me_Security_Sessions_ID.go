package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func DELETE_Users_Me_Security_Sessions_ID(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}
	ok, snowflake := tools.GetSnowflake(w, r)
	if !ok {
		return
	}

	// Delete Relevant Session
	tag, err := tools.Database.ExecContext(r.Context(),
		"DELETE FROM user_session WHERE id = ?",
		snowflake,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	if c, err := tag.RowsAffected(); err != nil {
		tools.SendServerError(w, r, err)
		return
	} else if c == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_SESSION)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
