package forum

import (
	"database/sql"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func dataEditUsername(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	dbstring := "./database.db"
	db, err := sql.Open("sqlite3", dbstring)

	_, err = db.Exec("UPDATE users SET username=? WHERE email=?", userInfos.EditedUsername, userInfos.Email)
	if err != nil {
		panic(err)
	}
	userInfos.Username = userInfos.EditedUsername
}

func dataEditEmail(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	dbstring := "./database.db"
	db, err := sql.Open("sqlite3", dbstring)

	_, err = db.Exec("UPDATE users SET email=? WHERE email=?", userInfos.EditedEmail, userInfos.Email)
	if err != nil {
		panic(err)
	}
	userInfos.Email = userInfos.EditedEmail
}

func dataEditPassword(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	dbstring := "./database.db"
	db, err := sql.Open("sqlite3", dbstring)

	if userInfos.EditedPassword != userInfos.ConfEditedPassword {
		userInfos.AccountError = "Les nouveaux mots de passe ne correspondent pas. Veuillez réessayer."
		userInfos.EditedPassword = ""
		userInfos.ConfEditedPassword = ""
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	editedpassword_hash, err := bcrypt.GenerateFromPassword([]byte(userInfos.EditedPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("UPDATE users SET password_hash=? WHERE email=?", string(editedpassword_hash), userInfos.Email)
	if err != nil {
		panic(err)
	}
	editedpassword_hash = nil
	userInfos.EditedPassword = ""
	userInfos.AccountError = "Mot de passe modifié avec succès."
}

func dataDeleteAccount(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	dbstring := "./database.db"
	db, err := sql.Open("sqlite3", dbstring)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	var comparepassword_hash string
	db.QueryRow("SELECT password_hash FROM users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.DeleteAccountPassword)) == nil {
		_, err = db.Exec("DELETE FROM users WHERE email=?", userInfos.Email)
		if err != nil {
			panic(err)
		}
		userInfos.AccountError = "Compte supprimé avec succès."
	} else {
		userInfos.AccountError = "Mot de passe incorrect. Impossible de supprimer le compte. Veuillez réessayer."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
	comparepassword_hash = ""
	userInfos.DeleteAccountPassword = ""
}
