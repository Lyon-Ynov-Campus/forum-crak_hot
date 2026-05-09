package forum

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func dataRegisterSend(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
	checkUsernameChar := "!\"#$%&'()*+,-./:;<=>?@[\\]^ ` {|}~€£¥©®™§"
	var usernameCharIsOk = true

	for _, charac := range userInfos.Username {
		if strings.ContainsRune(checkUsernameChar, charac) {
			usernameCharIsOk = false
			break
		}
	}

	if !usernameCharIsOk {
		userInfos.AccountError = "Le seul caractère spécial autorisé pour le nom d'utilisateur est _ . Veuillez réessayer."
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	checkPasswordCharacters(userInfos)
	if userInfos.AccountError != "" {
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if len(userInfos.Password) < 12 {
		userInfos.AccountError = "La taille du mot de passe doit être d'au moins 12 caractères."
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if userInfos.Password != userInfos.ConfPassword {
		userInfos.AccountError = "Les mots de passe ne correspondent pas."
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	password_hash, err := bcrypt.GenerateFromPassword([]byte(userInfos.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	res, err := db.Exec("INSERT INTO Users (username, email, password_hash) VALUES (?, ?, ?)", userInfos.Username, userInfos.Email, string(password_hash))
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "email") && strings.Contains(errMsg, "unique") {
			userInfos.AccountError = "Cet email est déjà utilisé."
		} else if strings.Contains(errMsg, "username") && strings.Contains(errMsg, "unique") {
			userInfos.AccountError = "Ce nom d'utilisateur est déjà utilisé."
		} else {
			userInfos.AccountError = errMsg
		}
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	lastID, _ := res.LastInsertId()
	token := GenerateToken(userInfos.Email)
	db.Exec("INSERT INTO Session (user_id, token) VALUES (?, ?)", lastID, token)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	db.QueryRow("SELECT id FROM Users WHERE email=?", userInfos.Email).Scan(&userInfos.DBid)
	userInfos.Password = ""
	userInfos.EditedPassword = ""
	userInfos.AccountError = ""
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func dataLoginCheck(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	var comparepassword_hash string
	var userID string

	if strings.Contains(userInfos.Email_Username, "@") {
		userInfos.Email = userInfos.Email_Username
		db.QueryRow("SELECT id, username, password_hash FROM Users WHERE email=?", userInfos.Email).Scan(&userID, &userInfos.Username, &comparepassword_hash)
	} else {
		userInfos.Username = userInfos.Email_Username
		db.QueryRow("SELECT id, email, password_hash FROM Users WHERE username=?", userInfos.Username).Scan(&userID, &userInfos.Email, &comparepassword_hash)
	}
	userInfos.DBid = userID

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.Password)) == nil {
		userInfos.AccountError = ""
		sessionToken := GenerateToken(userInfos.Email)

		db.Exec("DELETE FROM Session WHERE user_id = ?", userID)
		_, err := db.Exec("INSERT INTO Session (user_id, token) VALUES (?, ?)", userID, sessionToken)
		if err != nil {
			fmt.Println("Erreur création session DB:", err)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
		})

		userInfos.Password = ""
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	} else {
		userInfos.AccountError = "Email ou mot de passe incorrect."
		userInfos.Password = ""
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
