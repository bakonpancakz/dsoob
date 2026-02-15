package tools

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"time"
)

// Generate a Base32 Encoded TOTP Secret
func GenerateTOTPSecret() string {
	secret := make([]byte, 20)
	rand.Read(secret)
	return base32.StdEncoding.
		WithPadding(base32.NoPadding).
		EncodeToString(secret)
}

// Generate a URI usable by Authenticator Apps for Setup
func GenerateTOTPURI(issuer, username, secret string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s?period=30&digits=6&algorithm=SHA1&secret=%s&issuer=%s",
		url.PathEscape(username),
		url.PathEscape(secret),
		url.PathEscape(issuer),
	)
}

// Generate 6-digit TOTP Code for the Given Time
func GenerateTOTPCode(secret string, at time.Time) string {

	// Decode Secret
	key, err := base32.StdEncoding.WithPadding(base32.StdPadding).DecodeString(secret)
	if err != nil {
		panic(err)
	}

	// Generate Hash
	counter := uint64(at.Unix() / 30)
	message := make([]byte, 8)
	binary.BigEndian.PutUint64(message, counter)

	hash := hmac.New(sha1.New, key)
	hash.Write(message)
	sum := hash.Sum(nil)

	// Hash Truncation
	offset := sum[len(sum)-1] & 0x0F
	code := (int(sum[offset])&0x7F)<<24 | (int(sum[offset+1])&0xFF)<<16 | (int(sum[offset+2])&0xFF)<<8 | (int(sum[offset+3]) & 0xFF)
	return fmt.Sprintf("%06d", code%1000000)
}

// Checks a 6-digit TOTP Code against the Secret
func ValidateTOTPCode(code, secret string) bool {
	now := time.Now()
	for i := -1; i <= 1; i++ {
		timestep := now.Add(time.Duration(i*30) * time.Second)
		expected := GenerateTOTPCode(secret, timestep)
		if CompareStringConstant(code, expected) {
			return true
		}
	}
	return false
}
