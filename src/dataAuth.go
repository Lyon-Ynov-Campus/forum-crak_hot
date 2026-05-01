package forum

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"

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
		fmt.Println()
	}
}

func dataRegisterSend(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.AccountError = ""
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

	dbstring := "./database.db"
	db, err := sql.Open("sqlite3", dbstring)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	password_hash, err := bcrypt.GenerateFromPassword([]byte(userInfos.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	createUserTable := `
		CREATE TABLE IF NOT EXISTS users (
			id integer not null primary key autoincrement,
			email text not null unique,
			username text not null,
			password_hash text not null,
			status integer default 1,
			wins integer default 0,
			gamesplayed integer default 0
		);`
	_, err = db.Exec(createUserTable)

	senddata := `
		INSERT INTO users (email, username, password_hash) VALUES ('` + userInfos.Email + `', '` + userInfos.Username + `', '` + string(password_hash) + `');
		`
	_, err = db.Exec(senddata)
	if err != nil {
		userInfos.AccountError = "Cet email est déjà utilisé. Veuillez réessayer."
		http.Redirect(w, r, "/register", http.StatusSeeOther)
	} else {
		userInfos.AccountError = ""
		db.QueryRow("SELECT id FROM users WHERE email=?", userInfos.Email).Scan(&userInfos.DBid)
		http.SetCookie(w, &http.Cookie{
			Name:  "DBid",
			Value: userInfos.DBid,
		})
		userInfos.DBid = ""
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	}
	userInfos.Password = ""
	userInfos.EditedPassword = ""
	password_hash = nil
}

func dataLoginCheck(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	dbstring := "./database.db"
	db, err := sql.Open("sqlite3", dbstring)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	var comparepassword_hash string
	db.QueryRow("SELECT password_hash FROM users WHERE email=?", userInfos.Email).Scan(&comparepassword_hash)
	db.QueryRow("SELECT username FROM users WHERE email=?", userInfos.Email).Scan(&userInfos.Username)

	if bcrypt.CompareHashAndPassword([]byte(comparepassword_hash), []byte(userInfos.Password)) == nil {
		userInfos.AccountError = ""
		db.QueryRow("SELECT id FROM users WHERE email=?", userInfos.Email).Scan(&userInfos.DBid)
		http.SetCookie(w, &http.Cookie{
			Name:  "DBid",
			Value: userInfos.DBid,
		})
		userInfos.DBid = ""
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	} else {
		userInfos.AccountError = "Email ou mot de passe incorrect. Veuillez réessayer."
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	userInfos.Password = ""
	userInfos.EditedPassword = ""
	comparepassword_hash = ""
}
