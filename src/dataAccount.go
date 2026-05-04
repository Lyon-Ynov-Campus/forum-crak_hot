package forum

import (
	"net/http"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// CORRECTION : Ajout de l'étoile * et du nom de variable pour le pointeur
func checkEditedPasswordCharacters(userInfos *UserInfos) {
	var allowedCharacters = regexp.MustCompile(`^[\x21-\x7E]+$`)
	var CPC_hasUpper = regexp.MustCompile(`[A-Z]`)
	var CPC_hasLower = regexp.MustCompile(`[a-z]`)
	var CPC_hasDigit = regexp.MustCompile(`[0-9]`)
	var hasSpecial = regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`)

	if !allowedCharacters.MatchString(userInfos.EditedPassword) ||
		!CPC_hasUpper.MatchString(userInfos.EditedPassword) ||
		!CPC_hasLower.MatchString(userInfos.EditedPassword) ||
		!CPC_hasDigit.MatchString(userInfos.EditedPassword) ||
		!hasSpecial.MatchString(userInfos.EditedPassword) {
		userInfos.AccountError = "La composition du mot de passe ne respecte pas les critères attendus. Veuillez réessayer."
	}
}

func dataEditUsername(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
	checkUsernameChar := "!\"#$%&'()*+,-./:;<=>?@[\\]^ ` {|}~€£¥©®™§"
	var usernameCharIsOk = true

	for _, charac := range userInfos.EditedUsername {
		if strings.ContainsRune(checkUsernameChar, charac) {
			usernameCharIsOk = false
			break
		}
	}

	if !usernameCharIsOk {
		userInfos.AccountError = "Le seul caractère spécial autorisé est _ . Veuillez réessayer."
		return
	}

	_, err := db.Exec("UPDATE Users SET username=? WHERE email=?", userInfos.EditedUsername, userInfos.Email)
	if err != nil {
		userInfos.AccountError = "Ce nom d'utilisateur est déjà utilisé."
		return
	}
	userInfos.Username = userInfos.EditedUsername
	userInfos.AccountError = "Nom d'utilisateur modifié avec succès."
}

func dataEditEmail(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	_, err := db.Exec("UPDATE Users SET email=? WHERE email=?", userInfos.EditedEmail, userInfos.Email)
	if err != nil {
		userInfos.AccountError = "Cet email est déjà utilisé."
		return
	}
	userInfos.Email = userInfos.EditedEmail
	userInfos.AccountError = "Email modifié avec succès."
}

func dataEditPassword(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
	checkEditedPasswordCharacters(userInfos)
	if userInfos.AccountError != "" {
		userInfos.EditedPassword = ""
		userInfos.ConfEditedPassword = ""
		return
	}

	if len(userInfos.EditedPassword) < 12 {
		userInfos.AccountError = "La taille du mot de passe doit être d'au moins 12 caractères."
		return
	}

	if userInfos.EditedPassword != userInfos.ConfEditedPassword {
		userInfos.AccountError = "Les nouveaux mots de passe ne correspondent pas."
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(userInfos.EditedPassword), bcrypt.DefaultCost)
	_, err := db.Exec("UPDATE Users SET password_hash=? WHERE email=?", string(hash), userInfos.Email)
	if err != nil {
		panic(err)
	}
	userInfos.EditedPassword = ""
	userInfos.ConfEditedPassword = ""
	userInfos.AccountError = "Mot de passe modifié avec succès."
}

func dataDeleteAccount(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	var comparepassword_hash string
	// On utilise la DB globale déjà ouverte par ton collègue
	err := db.QueryRow("SELECT password_hash FROM Users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)
	if err != nil {
		userInfos.AccountError = "Erreur lors de la récupération du compte."
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.DeleteAccountPassword)) == nil {
		_, err = db.Exec("DELETE FROM Users WHERE email=?", userInfos.Email)
		if err != nil {
			panic(err)
		}

		// CORRECTION : On déconnecte l'utilisateur en supprimant son cookie
		userInfos.Username, userInfos.Email = "", ""
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		userInfos.AccountError = "Compte supprimé avec succès."
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		userInfos.AccountError = "Mot de passe incorrect. Impossible de supprimer le compte."
		userInfos.DeleteAccountPassword = ""
	}
}
