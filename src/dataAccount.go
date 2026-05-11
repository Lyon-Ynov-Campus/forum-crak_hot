package forum

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func checkEditedPasswordCharacters(userInfos *UserInfos) {
	var allowedCharacters = regexp.MustCompile(`^[\x21-\x7E]+$`)
	var CPC_hasUpper = regexp.MustCompile(`[A-Z]`)
	var CPC_hasLower = regexp.MustCompile(`[a-z]`)
	var CPC_hasDigit = regexp.MustCompile(`[0-9]`)
	var hasSpecial = regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\]^_{|}~]`)

	if len(userInfos.EditedPassword) < 12 ||
		!allowedCharacters.MatchString(userInfos.EditedPassword) ||
		!CPC_hasUpper.MatchString(userInfos.EditedPassword) ||
		!CPC_hasLower.MatchString(userInfos.EditedPassword) ||
		!CPC_hasDigit.MatchString(userInfos.EditedPassword) ||
		!hasSpecial.MatchString(userInfos.EditedPassword) {
		userInfos.AccountError = "La composition du mot de passe ne respecte pas les critères (12 caractères, Maj, Min, Chiffre, Spécial)."
	}
}

func updateDBPhoto(username string, ppURL string) error {
	_, err := db.Exec("UPDATE Users SET photo_profil = ? WHERE username = ?", ppURL, username)
	return err
}

func dataEditUsername(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if userInfos.EditedUsername == "" {
		return
	}
	_, err := db.Exec("UPDATE Users SET username=? WHERE email=?", userInfos.EditedUsername, userInfos.Email)
	if err != nil {
		userInfos.AccountError = "Ce pseudo est déjà utilisé."
		return
	}
	userInfos.Username = userInfos.EditedUsername
	userInfos.AccountError = "Pseudo mis à jour."
}

func dataEditEmail(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if userInfos.EditedEmail == "" {
		return
	}
	_, err := db.Exec("UPDATE Users SET email=? WHERE username=?", userInfos.EditedEmail, userInfos.Username)
	if err != nil {
		userInfos.AccountError = "Cette adresse email est déjà utilisée."
		return
	}
	userInfos.Email = userInfos.EditedEmail
	userInfos.AccountError = "Email mis à jour."
}

func dataEditPassword(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""

	if userInfos.EditedPassword != userInfos.ConfEditedPassword {
		userInfos.AccountError = "Les mots de passe ne correspondent pas."
		return
	}

	checkEditedPasswordCharacters(userInfos)
	if userInfos.AccountError != "" {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userInfos.EditedPassword), bcrypt.DefaultCost)
	if err != nil {
		userInfos.AccountError = "Erreur lors du hachage du mot de passe."
		return
	}

	_, err = db.Exec("UPDATE Users SET password_hash=? WHERE email=?", string(hash), userInfos.Email)
	if err != nil {
		userInfos.AccountError = "Erreur lors de la mise à jour en base de données."
		return
	}

	userInfos.EditedPassword = ""
	userInfos.ConfEditedPassword = ""
	userInfos.AccountError = "Mot de passe modifié avec succès."
}

func dataDeleteAccount(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	var comparepassword_hash string
	err := db.QueryRow("SELECT password_hash FROM Users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)
	if err != nil {
		userInfos.AccountError = "Erreur lors de la récupération du compte."
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.DeleteAccountPassword)) != nil {
		userInfos.AccountError = "Mot de passe incorrect. Impossible de supprimer le compte."
		return
	}

	if userInfos.LoadedPP != "" && strings.Contains(userInfos.LoadedPP, "/static/pp/") {
		if !strings.Contains(userInfos.LoadedPP, "defaultPP.png") {
			oldFile := strings.TrimPrefix(userInfos.LoadedPP, "/")
			os.Remove(oldFile)
		}
	}

	_, err = db.Exec("DELETE FROM Users WHERE email=?", userInfos.Email)
	if err != nil {
		userInfos.AccountError = "Erreur lors de la suppression."
		return
	}

	userInfos.AccountError = "Compte supprimé avec succès."
}