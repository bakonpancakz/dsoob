package tools

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var (
	ERROR_GENERIC_SERVER              = APIError{Status: 500, Code: 0, Message: "Server Error"}
	ERROR_GENERIC_NOT_FOUND           = APIError{Status: 404, Code: 0, Message: "Endpoint Not Found"}
	ERROR_GENERIC_RATELIMIT           = APIError{Status: 429, Code: 0, Message: "Too Many Requests"}
	ERROR_GENERIC_UNAUTHORIZED        = APIError{Status: 401, Code: 0, Message: "Unauthorized"}
	ERROR_GENERIC_METHOD_NOT_ALLOWED  = APIError{Status: 405, Code: 0, Message: "Method Not Allowed"}
	ERROR_GENERIC_GZIP_REQUIRED       = APIError{Status: 400, Code: 0, Message: "Support for GZIP is required for this endpoint"}
	ERROR_BODY_EMPTY                  = APIError{Status: 411, Code: 0, Message: "Request Body is Empty"}
	ERROR_BODY_TOO_LARGE              = APIError{Status: 413, Code: 0, Message: "Request Body is Too Large"}
	ERROR_BODY_INVALID_TYPE           = APIError{Status: 400, Code: 0, Message: "Invalid Body Type"}
	ERROR_BODY_INVALID_DATA           = APIError{Status: 422, Code: 0, Message: "Invalid Body"}
	ERROR_BODY_INVALID_FIELD          = APIError{Status: 400, Code: 0, Message: "Invalid Body Field"}
	ERROR_UNKNOWN_USER                = APIError{Status: 404, Code: 1020, Message: "Unknown User"}
	ERROR_UNKNOWN_SESSION             = APIError{Status: 404, Code: 1040, Message: "Unknown Session"}
	ERROR_UNKNOWN_IMAGE               = APIError{Status: 404, Code: 1050, Message: "Unknown Image"}
	ERROR_IMAGE_UNSUPPORTED           = APIError{Status: 400, Code: 2010, Message: "Unsupported Image Format (Supports: WEBP, GIF, JPEG, PNG)"}
	ERROR_IMAGE_MALFORMED             = APIError{Status: 400, Code: 2020, Message: "Invalid or Malformed Image Data"}
	ERROR_LOGIN_INCORRECT             = APIError{Status: 401, Code: 4010, Message: "Incorrect Email or Password"}
	ERROR_LOGIN_ACCOUNT_DELETED       = APIError{Status: 401, Code: 4020, Message: "Account Deleted"}
	ERROR_LOGIN_PASSWORD_RESET        = APIError{Status: 401, Code: 4030, Message: "Account Locked. Please reset your password using 'Forgot Password?' on the login page"}
	ERROR_LOGIN_PASSWORD_ALREADY_USED = APIError{Status: 400, Code: 4040, Message: "Password Already Used"}
	ERROR_SIGNUP_DUPLICATE_USERNAME   = APIError{Status: 409, Code: 4050, Message: "Username is already in use"}
	ERROR_SIGNUP_DUPLICATE_EMAIL      = APIError{Status: 409, Code: 4060, Message: "Email Address is already in use"}
	ERROR_MFA_EMAIL_SENT              = APIError{Status: 403, Code: 5010, Message: "Email Sent"}
	ERROR_MFA_EMAIL_ALREADY_VERIFIED  = APIError{Status: 400, Code: 5020, Message: "Email Address already Verified"}
	ERROR_MFA_PASSCODE_REQUIRED       = APIError{Status: 403, Code: 5030, Message: "Authenticator Passcode Required"}
	ERROR_MFA_PASSCODE_INCORRECT      = APIError{Status: 401, Code: 5040, Message: "Authenticator Passcode Incorrect"}
	ERROR_MFA_RECOVERY_CODE_USED      = APIError{Status: 401, Code: 5050, Message: "Recovery Code Used"}
	ERROR_MFA_RECOVERY_CODE_INCORRECT = APIError{Status: 403, Code: 5060, Message: "Recovery Code Incorrect"}
	ERROR_MFA_ESCALATION_REQUIRED     = APIError{Status: 403, Code: 5070, Message: "Escalation Required"}
	ERROR_MFA_PASSWORD_INCORRECT      = APIError{Status: 401, Code: 5080, Message: "Incorrect Password"}
	ERROR_MFA_DISABLED                = APIError{Status: 412, Code: 5090, Message: "MFA is Disabled"}
	ERROR_MFA_SETUP_ALREADY           = APIError{Status: 400, Code: 5100, Message: "MFA is Already Setup"}
	ERROR_MFA_SETUP_NOT_INITIALIZED   = APIError{Status: 412, Code: 5110, Message: "MFA Setup not Started"}
)

// Cancel Request and Respond with an API Error
func SendClientError(w http.ResponseWriter, r *http.Request, e APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	fmt.Fprintf(w, `{"code":%d,"message":%q}`, e.Code, e.Message)
}

// Cancel Request and Respond with a Generic Server Error
func SendServerError(w http.ResponseWriter, r *http.Request, err error) {

	debugStack := strings.Split(string(debug.Stack()), "\n")
	for i, item := range debugStack {
		debugStack[i] = strings.ReplaceAll(item, "\t", "    ")
	}
	if len(debugStack) > 5 {
		debugStack = debugStack[5:] //skip header
	}

	reqHeader := make(map[string]string, len(r.Header))
	for key, header := range r.Header {
		reqHeader[key] = strings.Join(header, ", ")
	}

	LoggerHTTP.Data(ERROR, err.Error(), map[string]any{
		"request": map[string]any{
			"method":  r.Method,
			"url":     r.URL.String(),
			"headers": reqHeader,
			"session": r.Context().Value(SESSION_KEY),
		},
		"error": map[string]any{
			"raw":     err,
			"message": err.Error(),
			"stack":   debugStack,
		},
	})
	SendClientError(w, r, ERROR_GENERIC_SERVER)
}
