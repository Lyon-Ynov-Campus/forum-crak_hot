package forum

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func loginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/login.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page", http.StatusInternalServerError)
		return
	}

	resetSent := r.URL.Query().Get("reset_sent") == "true"

	data := struct {
		*UserInfos
		IsConnected bool
		ResetSent   bool
	}{
		UserInfos:   userInfos,
		IsConnected: IsConnected(r),
		ResetSent:   resetSent,
	}
	tmpl.ExecuteTemplate(w, "login.html", data)
}

func checkloginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userInfos.Email_Username = r.FormValue("email_Username")
	userInfos.Password = r.FormValue("password")

	dataLoginCheck(w, r, userInfos)
}

func registerHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/register.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "register.html", userInfos)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	userInfos.Username = ""
	userInfos.Email = ""
	userInfos.IsConnected = false
	userInfos.AccountError = "Déconnecté avec succès."

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func editpasswordHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	userInfos.EditedPassword = r.FormValue("editedpassword")
	userInfos.ConfEditedPassword = r.FormValue("confeditedpassword")

	dataEditPassword(w, r, userInfos)

	if userInfos.AccountError == "Mot de passe modifié avec succès." {
		logoutHandler(w, r, userInfos)
	} else {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
}

func addPPHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	
	file, header, err := r.FormFile("addPP")
	if err != nil {
		userInfos.AccountError = "Erreur lors de l'upload de l'image."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	defer file.Close()

	uploadDir := "./static/pp/"
	os.MkdirAll(uploadDir, os.ModePerm)

	filename := fmt.Sprintf("%s%s", userInfos.Username, filepath.Ext(header.Filename))
	out, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		http.Error(w, "Erreur lors de la création du fichier", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Erreur lors de la copie du fichier", http.StatusInternalServerError)
		return
	}

	pathForDB := "/static/pp/" + filename
	err = updateDBPhoto(userInfos.Username, pathForDB)
	if err == nil {
		userInfos.LoadedPP = pathForDB
	}

	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func deleteaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	userInfos.DeleteAccountPassword = r.FormValue("deleteaccountpassword")

	dataDeleteAccount(w, r, userInfos)

	if userInfos.AccountError == "Compte supprimé avec succès." {
		logoutHandler(w, r, userInfos)
	} else {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
}

func checkregisterHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
    dataRegisterSend(w, r, userInfos)
}

func editaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
    homeHandler(w, r, userInfos) 
}

func editusernameHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
    userInfos.EditedUsername = r.FormValue("editedusername")
    dataEditUsername(w, r, userInfos)
    http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func editemailHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
    userInfos.EditedEmail = r.FormValue("editemail")
    dataEditEmail(w, r, userInfos)
    http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func deletePPHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
    userInfos.LoadedPP = ""
    updateDBPhoto(userInfos.Username, "")
    http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}
