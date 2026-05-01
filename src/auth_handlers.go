package forum

import (
	"html/template"
	"log"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("./pages/authScreen.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, userInfos)
}

func checkloginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.Email = r.FormValue("email")
	userInfos.Password = r.FormValue("password")

	dataLoginCheck(w, r, userInfos)
}

func registerHandler(w http.ResponseWriter, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("./pages/register.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, userInfos)
}

func checkregisterHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.Username = r.FormValue("username")
	userInfos.Email = r.FormValue("email")
	userInfos.Password = r.FormValue("password")
	userInfos.ConfPassword = r.FormValue("confpassword")

	dataRegisterSend(w, r, userInfos)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.Username,
		userInfos.EditedUsername,
		userInfos.Email,
		userInfos.EditedEmail,
		userInfos.AccountError =
		"", "", "", "", "Déconnecté avec succès."
	http.SetCookie(w, &http.Cookie{
		Name:  "DBid",
		Value: userInfos.DBid,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("./pages/account.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, userInfos)
}

func editusernameHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedUsername = r.FormValue("editedusername")
	dataEditUsername(w, r, userInfos)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func editemailHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedEmail = r.FormValue("editedemail")
	dataEditEmail(w, r, userInfos)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func editpasswordHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedPassword = r.FormValue("editedpassword")
	userInfos.ConfEditedPassword = r.FormValue("confeditedpassword")
	dataEditPassword(w, r, userInfos)
	if userInfos.AccountError == "Mot de passe modifié avec succès." {
		userInfos.Username,
			userInfos.EditedUsername,
			userInfos.Email,
			userInfos.EditedEmail =
			"", "", "", ""
		http.SetCookie(w, &http.Cookie{
			Name:  "DBid",
			Value: userInfos.DBid,
		})
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	}

}

func deleteaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.DeleteAccountPassword = r.FormValue("deleteaccountpassword")
	dataDeleteAccount(w, r, userInfos)
	if userInfos.AccountError == "Compte supprimé avec succès." {
		userInfos.Username,
			userInfos.EditedUsername,
			userInfos.Email,
			userInfos.EditedEmail =
			"", "", "", ""
		http.SetCookie(w, &http.Cookie{
			Name:  "DBid",
			Value: userInfos.DBid,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
