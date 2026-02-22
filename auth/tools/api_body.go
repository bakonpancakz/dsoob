package tools

import (
	"compress/gzip"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	BodyValidator     = validator.New(validator.WithRequiredStructEnabled())
	REGEX_USERNAME    = regexp.MustCompile("^[a-zA-Z0-9_]{3,32}$")        //
	REGEX_PASSCODE    = regexp.MustCompile("^([0-9]{6}|[0-9ABCDEF]{8})$") //
	REGEX_HAS_SPECIAL = regexp.MustCompile(`\P{L}`)                       // non-letter Unicode
	REGEX_HAS_UPPER   = regexp.MustCompile(`\p{Lu}`)                      // uppercase letter (any script)
	REGEX_HAS_LOWER   = regexp.MustCompile(`\p{Ll}`)                      // lowercase letter (any script)
	REGEX_HAS_NUMBER  = regexp.MustCompile(`[0-9]`)                       // numbers
)

func init() {

	BodyValidator.RegisterValidation("publickey", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		dec, err := base64.RawURLEncoding.DecodeString(str)
		if err != nil {
			return false
		}
		return len(dec) == ed25519.PublicKeySize
	})

	BodyValidator.RegisterValidation("token", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		return CompareTokenString(str)
	})

	BodyValidator.RegisterValidation("passcode", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		trm := strings.TrimSpace(str)
		if len(str) != len(trm) || !REGEX_PASSCODE.MatchString(str) {
			return false
		}
		return true
	})

	BodyValidator.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		// 64 characters due to bcrypt limitation
		if len(str) < 8 || len(str) > 64 || !REGEX_HAS_SPECIAL.MatchString(str) ||
			!REGEX_HAS_UPPER.MatchString(str) ||
			!REGEX_HAS_LOWER.MatchString(str) ||
			!REGEX_HAS_NUMBER.MatchString(str) {
			return false
		}
		return true
	})

	BodyValidator.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		trm := strings.TrimSpace(str)
		if len(str) != len(trm) || len(str) < 3 || len(str) > 32 || !REGEX_USERNAME.MatchString(str) {
			return false
		}
		return true
	})

	// Additionally displayname is used for subtitle validation
	BodyValidator.RegisterValidation("displayname", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		trm := strings.TrimSpace(str)
		if len(str) != len(trm) || len(str) > 32 {
			return false
		}
		return true
	})

	BodyValidator.RegisterValidation("description", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		trm := strings.TrimSpace(str)
		if len(str) != len(trm) || len(str) > 4096 {
			return false
		}
		return true
	})

	BodyValidator.RegisterValidation("color", func(fl validator.FieldLevel) bool {
		v := fl.Field().Int()
		if v < 0 || v > 16777215 {
			return false
		}
		return true
	})
}

// Decode Incoming JSON Request
func BindJSON(w http.ResponseWriter, r *http.Request, b any) bool {

	// Header Validation
	header := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(header, "application/json") {
		SendClientError(w, r, ERROR_BODY_INVALID_TYPE)
		return false
	}
	defer r.Body.Close()

	// Decode into Struct
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(b); err != nil {
		SendClientError(w, r, ERROR_BODY_INVALID_DATA)
		return false
	}

	// Struct Validation
	if err := BodyValidator.Struct(b); err != nil {
		fmt.Println("DELETE ME", err)
		SendClientError(w, r, ERROR_BODY_INVALID_FIELD)
		return false
	}

	return true
}

// Encode and Compress Outgoing Body
func SendJSON(w http.ResponseWriter, r *http.Request, s int, b any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Setup Compression
	var wr io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		wr = gz
	}
	w.WriteHeader(s)

	// Encode Content
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	return enc.Encode(b)
}
