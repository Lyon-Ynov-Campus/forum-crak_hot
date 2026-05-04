package forum

import (
	"html/template"
	"log"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/login.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Fatal(err)
	}
	data := struct {
		*UserInfos
		IsConnected bool
	}{
		UserInfos:   userInfos,
		IsConnected: IsConnected(r),
	}
	tmpl.ExecuteTemplate(w, "login.html", data)
}

func checkloginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.Email_Username = r.FormValue("email_Username")
	userInfos.Password = r.FormValue("password")
	dataLoginCheck(w, r, userInfos)
}

func registerHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/register.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur template register : %v", err)
		return
	}
	data := struct {
		*UserInfos
		IsConnected bool
	}{
		UserInfos:   userInfos,
		IsConnected: IsConnected(r),
	}

	tmpl.ExecuteTemplate(w, "register.html", data)
}

func checkregisterHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.Username = r.FormValue("username")
	userInfos.Email = r.FormValue("email")
	userInfos.Password = r.FormValue("password")
	userInfos.ConfPassword = r.FormValue("confpassword")
	dataRegisterSend(w, r, userInfos)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.Username, userInfos.EditedUsername, userInfos.Email, userInfos.EditedEmail, userInfos.AccountError = "", "", "", "", "Déconnecté avec succès."
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur template account : %v", err)
		return
	}
	data := struct {
		*UserInfos
		IsConnected bool
	}{
		UserInfos:   userInfos,
		IsConnected: IsConnected(r),
	}

	tmpl.ExecuteTemplate(w, "account.html", data)
}

func editusernameHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedUsername = r.FormValue("editedusername")
	dataEditUsername(w, r, userInfos)
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func editemailHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedEmail = r.FormValue("editedemail")
	dataEditEmail(w, r, userInfos)
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func editpasswordHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedPassword = r.FormValue("editedpassword")
	userInfos.ConfEditedPassword = r.FormValue("confeditedpassword")
	dataEditPassword(w, r, userInfos)

	if userInfos.AccountError == "Mot de passe modifié avec succès." {
		userInfos.Username, userInfos.Email = "", ""
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
}

func deleteaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.DeleteAccountPassword = r.FormValue("deleteaccountpassword")
	dataDeleteAccount(w, r, userInfos)
	if userInfos.AccountError == "Compte supprimé avec succès." {
		userInfos.Username, userInfos.Email = "", ""
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
