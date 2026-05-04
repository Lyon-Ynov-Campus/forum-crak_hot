package forum

import (
	"encoding/base64"
	"fmt"
	"time"
)

func GenerateToken(pseudo string) string {
	expiration := time.Now().Add(1 * time.Hour)
	data := fmt.Sprint("%s|%s", pseudo, expiration)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func ValidateToken(tokenStr string) (string, bool) {
	data, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return "", false
	}
	return string(data), true
}
