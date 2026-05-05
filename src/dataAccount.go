package forum

import (
	"database/sql"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func checkEditedPasswordCharacters(UserInfos) {
	var allowedCharacters = regexp.MustCompile(`^[\x21-\x7E]+$`)
	var CPC_hasUpper = regexp.MustCompile(`[A-Z]`)
	var CPC_hasLower = regexp.MustCompile(`[a-z]`)
	var CPC_hasDigit = regexp.MustCompile(`[0-9]`)
	var hasSpecial = regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`)

	if !allowedCharacters.MatchString(userInfos.EditedPassword) || !CPC_hasUpper.MatchString(userInfos.EditedPassword) || !CPC_hasLower.MatchString(userInfos.EditedPassword) || !CPC_hasDigit.MatchString(userInfos.EditedPassword) || !hasSpecial.MatchString(userInfos.EditedPassword) {
		userInfos.AccountError = "La composition du mot de passe de respecte pas les critères attendus. Veuillez réessayer."
	}
}

func updateUserPP(email string, ppURL string) error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE Users SET photo_profil = ? WHERE email = ?", ppURL, email)
	return err
}

func getUserPP(email string) (string, error) {
	db, err := OpenDB()
	if err != nil {
		return "", err
	}

	var ppURL string
	err = db.QueryRow("SELECT photo_profil FROM Users WHERE email = ?", email).Scan(&ppURL)
	if err != nil {
		return "", err
	}
	return ppURL, nil
}

func dataEditUsername(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
	checkUsernameChar := "!\"#$%&'()*+,-./:;<=>?@[\\]^ ` {|}~€£¥©®™§"
	var usernameCharIsOk bool

	for _, charac := range userInfos.EditedUsername {
		if strings.ContainsRune(checkUsernameChar, charac) {
			usernameCharIsOk = false
			break
		} else {
			usernameCharIsOk = true
		}
	}

	if usernameCharIsOk == false {
		userInfos.AccountError = "Le seul caractères spécial autorisé du nom d'utilisateur est _ . Veuillez réessayer."
		return
	}

	_, err := db.Exec("UPDATE Users SET username=? WHERE email=?", userInfos.EditedUsername, userInfos.Email)
	if err != nil {
		errMsg := err.Error()
		if regexp.MustCompile(`(?i)username`).MatchString(errMsg) && regexp.MustCompile(`(?i)unique`).MatchString(errMsg) {
			userInfos.AccountError = "Ce nom d'utilisateur est déjà utilisé. Veuillez en choisir un autre."
			return
		} else {
			userInfos.AccountError = errMsg
			return
		}
	}
	userInfos.Username = userInfos.EditedUsername
	userInfos.AccountError = "Nom d'utilisateur modifié avec succès."
}

func dataEditEmail(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	_, err := db.Exec("UPDATE Users SET email=? WHERE email=?", userInfos.EditedEmail, userInfos.Email)
	if err != nil {
		errMsg := err.Error()
		if regexp.MustCompile(`(?i)email`).MatchString(errMsg) && regexp.MustCompile(`(?i)unique`).MatchString(errMsg) {
			userInfos.AccountError = "Cet email est déjà utilisé. Veuillez en choisir un autre."
			return
		} else {
			userInfos.AccountError = errMsg
			return
		}
	}
	userInfos.Email = userInfos.EditedEmail
	userInfos.AccountError = "Email modifié avec succès."
}

func dataEditPassword(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
	checkEditedPasswordCharacters(*userInfos)
	if userInfos.AccountError == "La composition du mot de passe de respecte pas les critères attendus. Veuillez réessayer." {
		userInfos.EditedPassword = ""
		userInfos.ConfEditedPassword = ""
		return
	}

	if len(userInfos.EditedPassword) < 12 {
		userInfos.AccountError = "La taille du mot de passe doit être d'au moins 12 caracères. Veuillez réessayer."
		userInfos.EditedPassword = ""
		userInfos.ConfEditedPassword = ""
		return
	}

	if userInfos.EditedPassword != userInfos.ConfEditedPassword {
		userInfos.AccountError = "Les nouveaux mots de passe ne correspondent pas. Veuillez réessayer."
		userInfos.EditedPassword = ""
		userInfos.ConfEditedPassword = ""
		return
	}

	editedpassword_hash, err := bcrypt.GenerateFromPassword([]byte(userInfos.EditedPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("UPDATE Users SET password_hash=? WHERE email=?", string(editedpassword_hash), userInfos.Email)
	if err != nil {
		panic(err)
	}
	editedpassword_hash = nil
	userInfos.EditedPassword = ""
	userInfos.AccountError = "Mot de passe modifié avec succès."
}

func dataDeleteAccount(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	dbstring := "./Forum.db"
	db, err := sql.Open("sqlite3", dbstring)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	var comparepassword_hash string
	db.QueryRow("SELECT password_hash FROM Users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.DeleteAccountPassword)) == nil {
		// Supprime la photo de profil si elle existe
		if userInfos.LoadedPP != "" && strings.Contains(userInfos.LoadedPP, "/static/pp/") {
			parts := strings.Split(userInfos.LoadedPP, "/static/pp/")
			if len(parts) == 2 {
				oldFile := "static/pp/" + parts[1]
				os.Remove(oldFile)
			}
		}

		_, err = db.Exec("DELETE FROM Users WHERE email=?", userInfos.Email)
		if err != nil {
			panic(err)
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "DBid",
			Value: "",
		})
		userInfos.AccountError = "Compte supprimé avec succès."
	} else {
		userInfos.AccountError = "Mot de passe incorrect. Impossible de supprimer le compte. Veuillez réessayer."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
	comparepassword_hash = ""
	userInfos.DeleteAccountPassword = ""
}
