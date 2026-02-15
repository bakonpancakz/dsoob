package tools

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path"
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
	CONTEXT_TIMEOUT                          = 10 * time.Second    // Default Context Timeout
	SHUTDOWN_TIMEOUT                         = 1 * time.Minute     // Default Shutdown Timeout
	LIFETIME_METADATA_CACHE                  = 10 * time.Minute    // Lifetime for Metadata File
	LIFETIME_TOKEN_USER_ELEVATION            = 10 * time.Minute    // Lifetime for User Elevation
	LIFETIME_TOKEN_USER_COOKIE               = 30 * 24 * time.Hour // Lifetime for User Cookie
	LIFETIME_TOKEN_EMAIL_PASSCODE            = 15 * time.Minute    // Lifetime for MFA Passcode
	LIFETIME_TOKEN_EMAIL_LOGIN               = 24 * time.Hour      // Lifetime for Verify Login Token
	LIFETIME_TOKEN_EMAIL_VERIFY              = 24 * time.Hour      // Lifetime for Verify Email Token
	LIFETIME_TOKEN_EMAIL_RESET               = 24 * time.Hour      // Lifetime for Password Reset Token
	PASSWORD_HASH_EFFORT                     = 12                  // Password Hashing Effort
	PASSWORD_HISTORY_LIMIT                   = 5                   // Password History Length
	MFA_PASSCODE_LENGTH                      = 6                   // TOTP Passcode String Length (Do Not Change)
	MFA_RECOVERY_LENGTH                      = 8                   // TOTP Recovery Code Length (Do Not Change)
	TOKEN_PREFIX_USER                        = "User "
	SESSION_KEY                   contextKey = "gloopert"
)

var (
	DATA_DIRECTORY        = EnvString("DATA_DIRECTORY", "./data")
	EMAIL_SMTP_HOST       = EnvString("EMAIL_SMTP_HOST", "127.0.0.1")
	EMAIL_SMTP_ADDRESS    = EnvString("EMAIL_SMTP_ADDRESS", "noreply@example.org")
	STORAGE_S3_KEY_SECRET = EnvString("STORAGE_S3_KEY_SECRET", "xyz")
	STORAGE_S3_KEY_ACCESS = EnvString("STORAGE_S3_KEY_ACCESS", "123")
	STORAGE_S3_ENDPOINT   = EnvString("STORAGE_S3_ENDPOINT", "https://bucket.s3.region.host.tld")
	STORAGE_S3_REGION     = EnvString("STORAGE_S3_REGION", "region")
	STORAGE_S3_BUCKET     = EnvString("STORAGE_S3_BUCKET", "bucket")
	HTTP_ADDRESS          = EnvString("HTTP_ADDRESS", "127.0.0.1:8080")
	HTTP_IP_HEADERS       = EnvSlice("HTTP_IP_HEADERS", ",", []string{"X-Forwarded-For"})
	HTTP_IP_PROXIES       = EnvSlice("HTTP_IP_PROXIES", ",", []string{"127.0.0.1/8"})
	HTTP_TLS_ENABLED      = EnvString("HTTP_TLS_ENABLED", "false") == "true"
	HTTP_TLS_CERT         = EnvString("HTTP_TLS_CERT", "tls_crt.pem")
	HTTP_TLS_KEY          = EnvString("HTTP_TLS_KEY", "tls_key.pem")
	HTTP_TLS_CA           = EnvString("HTTP_TLS_CA", "tls_ca.pem")
	HTTP_KEY              = []byte(EnvString("HTTP_KEY", "teto"))
)

func init() {
	// Create Directories
	// 	This will cause fatal errors elsewhere so it's ok!
	os.MkdirAll(path.Join(DATA_DIRECTORY), FILEMODE_PUBLIC)
	os.MkdirAll(path.Join(DATA_DIRECTORY, "settings"), FILEMODE_SECURE)
	os.MkdirAll(path.Join(DATA_DIRECTORY, "database"), FILEMODE_SECURE)
}

// Default Context Timeout
func NewContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
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
func EnvString(field, initial string) string {
	if value := os.Getenv(field); value == "" {
		return initial
	} else {
		return value
	}
}

// Read String from Environment and Parse it as a slice using the given delimiter
func EnvSlice(field, delimiter string, initial []string) []string {
	if value := os.Getenv(field); value == "" {
		return initial
	} else {
		return strings.Split(value, delimiter)
	}
}
