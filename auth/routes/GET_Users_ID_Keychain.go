package routes

import (
	"dsoob/backend/tools"
	"net/http"
)

func GET_Users_ID_Keychain(w http.ResponseWriter, r *http.Request) {

	ok, userID := tools.GetSnowflake(w, r)
	if !ok {
		return
	}

	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Sessions
	rows, err := tools.Database.QueryContext(ctx,
		"SELECT device_public_key FROM user_session WHERE user_id = $1",
		userID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	defer rows.Close()

	// Organize Sessions
	var (
		UserItems = make([]string, 0, 8)
		UserKey   string
	)
	for rows.Next() {
		if err := rows.Scan(&UserKey); err != nil {
			tools.SendServerError(w, r, err)
			return
		}
	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, UserItems)
}
