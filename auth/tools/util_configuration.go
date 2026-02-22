package tools

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type contextKey string

const (
	FILEMODE_PUBLIC                          = os.FileMode(0770)   // rw-rw----
	FILEMODE_SECURE                          = os.FileMode(0700)   // rw-------
	ARRAY_DELIMITER                          = "|"                 // For Database arrays, do not use for any user input!
	SITE_NAME                                = "dsoob.net"         // Used for Emails and TOTP
	EPOCH_MILLI                              = 1207008000000       // Generic EPOCH (April 1st 2008, Teto b-day!)
	EPOCH_SECONDS                            = EPOCH_MILLI / 1000  // Generic EPOCH in Seconds
	TIMEOUT_SHUTDOWN                         = 1 * time.Minute     // Default Timeout for Shutdowns
	TIMEOUT_CONTEXT                          = 10 * time.Second    // Default Timeout for Requests
	PASSWORD_HASH_EFFORT                     = 12                  // Password Hashing Effort
	PASSWORD_HISTORY_LIMIT                   = 5                   // Password History Length
	PASSWORD_CONCURRENT_LIMIT                = 8                   // Password Hashing Concurrency Limit
	MFA_PASSCODE_LENGTH                      = 6                   // TOTP Passcode String Length (Do Not Change)
	MFA_RECOVERY_LENGTH                      = 8                   // TOTP Recovery Code Length (Do Not Change)
	TOKEN_LIFETIME_USER_ELEVATION            = 10 * time.Minute    // Lifetime for User Elevation
	TOKEN_LIFETIME_USER_COOKIE               = 30 * 24 * time.Hour // Lifetime for User Cookie
	TOKEN_LIFETIME_EMAIL_PASSCODE            = 15 * time.Minute    // Lifetime for MFA Passcode
	TOKEN_LIFETIME_EMAIL_LOGIN               = 24 * time.Hour      // Lifetime for Verify Login Token
	TOKEN_LIFETIME_EMAIL_VERIFY              = 24 * time.Hour      // Lifetime for Verify Email Token
	TOKEN_LIFETIME_EMAIL_RESET               = 24 * time.Hour      // Lifetime for Password Reset Token
	TOKEN_BYTE_LENGTH                        = 64
	TOKEN_PREFIX_USER                        = "User "
	SESSION_KEY                   contextKey = "gloopert"
)

var (
	DATA_DIRECTORY     = envString("DATA_DIRECTORY", "./data")
	EMAIL_SMTP_HOST    = envString("EMAIL_SMTP_HOST", "127.0.0.1:1273")
	EMAIL_SMTP_ADDRESS = envString("EMAIL_SMTP_ADDRESS", "noreply@example.org")
	HTTP_ADDRESS       = envString("HTTP_ADDRESS", "127.0.0.1:8080")
	HTTP_IP_HEADERS    = envSlice("HTTP_IP_HEADERS", ",", []string{"X-Forwarded-For"})
	HTTP_IP_PROXIES    = envSlice("HTTP_IP_PROXIES", ",", []string{"127.0.0.1/8"})
	HTTP_TLS_ENABLED   = envString("HTTP_TLS_ENABLED", "false") == "true"
	HTTP_TLS_CERT      = envString("HTTP_TLS_CERT", "tls_crt.pem")
	HTTP_TLS_KEY       = envString("HTTP_TLS_KEY", "tls_key.pem")
	HTTP_TLS_CA        = envString("HTTP_TLS_CA", "tls_ca.pem")
)

func init() {
	// Prepare Data Directories
	for _, item := range []struct {
		Permissions os.FileMode
		Directory   string
	}{
		{FILEMODE_PUBLIC, "public"},
		{FILEMODE_SECURE, "settings"},
		{FILEMODE_SECURE, "database"},
	} {
		// Attempt to Create Directory
		pth := path.Join(DATA_DIRECTORY, item.Directory)
		err := os.MkdirAll(pth, item.Permissions)
		if err != nil {
			LoggerMain.Log(FATAL, "Cannot create '%s' directory: %s", pth, err)
			return
		}

		// Override Permissions
		err = os.Chmod(pth, item.Permissions)
		if err != nil {
			LoggerMain.Log(FATAL, "Cannot gatekeep '%s' directory: %s", pth, err)
			return
		}
	}
}

// Create TLS Configuration from Crypto
func NewTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(HTTP_TLS_CERT, HTTP_TLS_KEY)
	if err != nil {
		return nil, err
	}
	caBytes, err := os.ReadFile(HTTP_TLS_CA)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caBytes) {
		return nil, errors.New("cannot append ca bundle")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}, nil
}

// Read String from Environment
func envString(field, initial string) string {
	if value := os.Getenv(field); value == "" {
		return initial
	} else {
		return value
	}
}

// Read String from Environment and Parse it as a slice using the given delimiter
func envSlice(field, delimiter string, initial []string) []string {
	if value := os.Getenv(field); value == "" {
		return initial
	} else {
		return strings.Split(value, delimiter)
	}
}

// Read String from Environment and Parse it as a number
func envNumber(field string, initial int) int {
	if value := os.Getenv(field); value == "" {
		return initial
	} else if number, err := strconv.Atoi(value); err != nil {
		return initial
	} else {
		return number
	}
}
