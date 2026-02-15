package tools

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// Applies relevant headers for AWS Signature Version 4
func AmazonSignRequestV4(req *http.Request, payload []byte, accessKey, secretKey, host, region, service string) {
	// Timestamp
	t := time.Now().UTC()
	dateStamp := t.Format("20060102")
	dataAmazon := t.Format("20060102T150405Z")

	// Hash Content
	sum := md5.Sum(payload)
	payloadHashHexSHA := sha256HEX(payload)
	payloadHashB64MD5 := base64.StdEncoding.EncodeToString(sum[:])

	// Apply Headers
	req.Header.Set("Host", host)
	req.Header.Set("Content-MD5", payloadHashB64MD5)
	req.Header.Set("x-amz-date", dataAmazon)
	req.Header.Set("x-amz-content-sha256", payloadHashHexSHA)
	req.Host = host

	// 1. Create a canonical request
	signedHeaders := "content-md5;host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := fmt.Sprintf(
		"content-md5:%s\nhost:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		payloadHashB64MD5, req.Host, payloadHashHexSHA, dataAmazon,
	)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.EscapedPath(),
		req.URL.Query().Encode(),
		canonicalHeaders,
		signedHeaders,
		payloadHashHexSHA,
	)

	// Step 2: String to sign
	scope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		dataAmazon, scope, sha256HEX([]byte(canonicalRequest)),
	)

	// Step 3: Derive signing key
	kDate := sha256HMAC([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := sha256HMAC(kDate, []byte(region))
	kService := sha256HMAC(kRegion, []byte(service))
	kSigning := sha256HMAC(kService, []byte("aws4_request"))

	// Step 4: Signature
	signature := hex.EncodeToString(sha256HMAC(kSigning, []byte(stringToSign)))
	authHeader := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, scope, signedHeaders, signature,
	)
	req.Header.Set("Authorization", authHeader)
}

func sha256HEX(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func sha256HMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
