package routes

import (
	"net/http"

	"dsoob/backend/tools"
)

func GET_Users_Me_Security_Sessions(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)

	// Fetch Sessions
	rows, err := tools.Database.QueryContext(r.Context(),
		"SELECT id, device_ip_address, device_user_agent, device_public_key FROM user_session WHERE user_id = $1",
		session.UserID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	defer rows.Close()

	// Organize Sessions
	var (
		LoginItems           = make([]map[string]any, 0, 1)
		LoginID              int64
		LoginDeviceIPAddress string
		LoginDeviceUserAgent string
		LoginDevicePublicKey string
	)
	for rows.Next() {
		if err := rows.Scan(
			&LoginID,
			&LoginDeviceIPAddress,
			&LoginDeviceUserAgent,
			&LoginDevicePublicKey,
		); err != nil {
			tools.SendServerError(w, r, err)
			return
		}
		LoginItems = append(LoginItems, map[string]any{
			"id":         LoginID,
			"location":   tools.LookupLocation(LoginDeviceIPAddress),
			"browser":    tools.LookupBrowser(LoginDeviceUserAgent),
			"public_key": LoginDevicePublicKey,
		})
	}

	// Return Results
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
		"current":  session.SessionID,
		"sessions": LoginItems,
	})
}
