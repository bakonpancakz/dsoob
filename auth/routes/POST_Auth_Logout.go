package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func POST_Auth_Logout(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)

	// Revoke Current Session
	tag, err := tools.Database.ExecContext(r.Context(),
		"DELETE FROM user_session WHERE id = ? AND user_id = ?",
		session.SessionID,
		session.UserID,
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
