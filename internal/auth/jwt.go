package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"uptime-monitor/internal/config"
)

var secretKey = config.LoadEnv("JWT_SECRET")

func GenerateJWT(userID int) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	payload := map[string]interface{}{
		"user_id": userID,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsignedToken := headerB64 + "." + payloadB64

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(unsignedToken))
	signatureBytes := mac.Sum(nil)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signatureBytes)

	token := unsignedToken + "." + signatureB64

	return token, nil

}

func Verify(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid error format")
	}
	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	unsignedToken := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(unsignedToken))
	expectedSignatureBytes := mac.Sum(nil)
	expectedSignatureB64 := base64.RawURLEncoding.EncodeToString(expectedSignatureBytes)
	if !hmac.Equal([]byte(expectedSignatureB64), []byte(signatureB64)) {
		return nil, errors.New("invalid signature / token has been tempered")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	json.Unmarshal(payloadBytes, &payload)

	expFloat, ok := payload["exp"].(float64)
	if ok {
		if time.Now().Unix() > int64(expFloat) {
			return nil, errors.New("token is expired")

		}
	}
	return payload, nil

}
