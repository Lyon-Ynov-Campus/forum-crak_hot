package forum

import (
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func checkPasswordCharacters(UserInfos) {
	var allowedCharacters = regexp.MustCompile(`^[\x21-\x7E]+$`)
	var CPC_hasUpper = regexp.MustCompile(`[A-Z]`)
	var CPC_hasLower = regexp.MustCompile(`[a-z]`)
	var CPC_hasDigit = regexp.MustCompile(`[0-9]`)
	var hasSpecial = regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`)

	if !allowedCharacters.MatchString(userInfos.Password) || !CPC_hasUpper.MatchString(userInfos.Password) || !CPC_hasLower.MatchString(userInfos.Password) || !CPC_hasDigit.MatchString(userInfos.Password) || !hasSpecial.MatchString(userInfos.Password) {
		userInfos.AccountError = "La composition du mot de passe de respecte pas les critères attendus. Veuillez réessayer."
	}
}

func dataRegisterSend(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
	checkUsernameChar := "!\"#$%&'()*+,-./:;<=>?@[\\]^ ` {|}~€£¥©®™§"
	var usernameCharIsOk bool

	for _, charac := range userInfos.Username {
		if strings.ContainsRune(checkUsernameChar, charac) {
			usernameCharIsOk = false
			break
		} else {
			usernameCharIsOk = true
		}
	}

	if usernameCharIsOk == false {
		userInfos.AccountError = "Le seul caractères spécial autorisé du nom d'utilisateur est _ . Veuillez réessayer."
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	checkPasswordCharacters(*userInfos)
	if userInfos.AccountError == "La composition du mot de passe de respecte pas les critères attendus. Veuillez réessayer." {
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if len(userInfos.Password) < 12 {
		userInfos.AccountError = "La taille du mot de passe doit être d'au moins 12 caracères. Veuillez réessayer."
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if userInfos.Password != userInfos.ConfPassword {
		userInfos.AccountError = "Les mots de passe ne correspondent pas. Veuillez réessayer."
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	password_hash, err := bcrypt.GenerateFromPassword([]byte(userInfos.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("INSERT INTO Users (username, email, password_hash) VALUES (?, ?, ?)", userInfos.Username, userInfos.Email, string(password_hash))
	if err != nil {
		errMsg := err.Error()
		if regexp.MustCompile(`(?i)email`).MatchString(errMsg) && regexp.MustCompile(`(?i)unique`).MatchString(errMsg) {
			userInfos.AccountError = "Cet email est déjà utilisé. Veuillez en choisir un autre."
		} else if regexp.MustCompile(`(?i)username`).MatchString(errMsg) && regexp.MustCompile(`(?i)unique`).MatchString(errMsg) {
			userInfos.AccountError = "Ce nom d'utilisateur est déjà utilisé. Veuillez en choisir un autre."
		} else {
			userInfos.AccountError = errMsg
		}
		userInfos.Password = ""
		userInfos.ConfPassword = ""
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	userInfos.AccountError = ""
	db.QueryRow("SELECT id FROM Users WHERE email=?", userInfos.Email).Scan(&userInfos.DBid)
	http.SetCookie(w, &http.Cookie{
		Name:  "DBid",
		Value: userInfos.DBid,
	})
	userInfos.DBid = ""
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
	userInfos.Password = ""
	userInfos.EditedPassword = ""
	password_hash = nil
}

func dataLoginCheck(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	var comparepassword_hash string

	if strings.Contains(userInfos.Email_Username, "@") {
		// Le champs de login est un mail
		userInfos.Email = userInfos.Email_Username

		db.QueryRow("SELECT username FROM Users WHERE email=?", userInfos.Email).Scan(&userInfos.Username)
		db.QueryRow("SELECT password_hash FROM Users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)
	} else {
		// Le champs de login est un username
		userInfos.Username = userInfos.Email_Username

		db.QueryRow("SELECT email FROM Users WHERE username=?", userInfos.Email_Username).Scan(&userInfos.Email)
		db.QueryRow("SELECT password_hash FROM Users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)
	}

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.Password)) == nil {
		userInfos.AccountError = ""
		db.QueryRow("SELECT id FROM Users WHERE email=?", userInfos.Email).Scan(&userInfos.DBid)
		http.SetCookie(w, &http.Cookie{
			Name:  "DBid",
			Value: userInfos.DBid,
		})
		userInfos.DBid = ""
		userInfos.Password = ""
		userInfos.EditedPassword = ""
		comparepassword_hash = ""
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	} else {
		userInfos.AccountError = "Email ou mot de passe incorrect. Veuillez réessayer."
		userInfos.Password = ""
		userInfos.EditedPassword = ""
		comparepassword_hash = ""
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

}
