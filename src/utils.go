package forum

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GenerateToken(pseudo string) string {
	expiration := time.Now().Add(1 * time.Hour)
	data := fmt.Sprintf("%s|%s", pseudo, expiration)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func ValidateToken(tokenStr string) (string, bool) {
	data, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func SetFlash(w http.ResponseWriter, kind, message string) {
	value := url.QueryEscape(kind + "|" + message)
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    value,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func GetFlash(w http.ResponseWriter, r *http.Request) (string, string) {
	cookie, err := r.Cookie("flash")
	if err != nil {
		return "", ""
	}
	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		ClearFlash(w)
		return "", ""
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		ClearFlash(w)
		return "", ""
	}
	ClearFlash(w)
	return parts[0], parts[1]
}

func ClearFlash(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func LoadFlash(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	kind, message := GetFlash(w, r)
	userInfos.FlashType = kind
	userInfos.FlashMessage = message
	userInfos.AccountError = ""
}

func PushAccountFlash(w http.ResponseWriter, userInfos *UserInfos) {
	if userInfos.AccountError == "" {
		return
	}
	kind := "error"
	switch userInfos.AccountError {
	case "Nom d'utilisateur modifié avec succès.", "Email modifié avec succès.", "Mot de passe modifié avec succès.", "Compte supprimé avec succès.", "Photo de profil mise à jour.", "Photo de profil supprimée.", "Déconnecté avec succès.":
		kind = "success"
	}
	SetFlash(w, kind, userInfos.AccountError)
	userInfos.AccountError = ""
}
