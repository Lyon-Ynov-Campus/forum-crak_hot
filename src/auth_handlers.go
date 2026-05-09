package forum

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func loginHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/login.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Fatal(err)
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
	user := GetUserFromSession(r)

	pp, err := getUserPP(user.Email)
	if err == nil {
		user.LoadedPP = pp
	} else {
		user.LoadedPP = ""
	}

	if !user.IsConnected {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID := GetUserID(r)

	posts, _ := GetUserPosts(userID)

	comments, _ := GetUserCom(userID)

	tmpl, err := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur template account : %v", err)
		return
	}

	data := struct {
		Posts    []Post
		Comments []Com
		*UserInfos
		Page string
	}{
		Posts:     posts,
		Comments:  comments,
		UserInfos: user,
		Page:      "account",
	}
	tmpl.ExecuteTemplate(w, "account.html", data)
}

func addPPHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if userInfos.Email == "" {
		log.Println("Tentative d'upload sans Email")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		log.Println("Erreur ParseMultipartForm:", err)
		userInfos.AccountError = "Erreur lors de l'upload."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	file, handler, err := r.FormFile("addPP")
	if err != nil {
		log.Println("Erreur FormFile:", err)
		userInfos.AccountError = "Erreur lors de la récupération du fichier."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		userInfos.AccountError = "Format non supporté."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	ppDir := filepath.Join("static", "pp")
	if _, err := os.Stat(ppDir); os.IsNotExist(err) {
		if err := os.MkdirAll(ppDir, 0755); err != nil {
			log.Println("Erreur création dossier static/pp:", err)
			userInfos.AccountError = "Erreur serveur (dossier)."
			http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
			return
		}
	}

	fileNamePart := "PPofNum" + userInfos.DBid
	filename := fmt.Sprintf("%s%s", fileNamePart, ext)
	path := filepath.Join(ppDir, filename)

	removeOldPP(userInfos)

	out, err := os.Create(path)
	if err != nil {
		log.Println("Erreur création fichier:", err)
		userInfos.AccountError = "Erreur lors de la sauvegarde."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("Erreur copie fichier:", err)
		userInfos.AccountError = "Erreur lors de la copie."
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	ppURL := "/static/pp/" + filename
	userInfos.EditedPP = ppURL
	userInfos.LoadedPP = ppURL

	err = updateUserPP(userInfos.Email, ppURL)
	if err != nil {
		log.Println("Erreur updateUserPP:", err)
		userInfos.AccountError = "Erreur lors de la mise à jour de la base de données."
	} else {
		userInfos.AccountError = "Photo de profil mise à jour."
	}
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func removeOldPP(userInfos *UserInfos) {
	if userInfos.LoadedPP != "" && strings.Contains(userInfos.LoadedPP, "/static/pp/") {
		parts := strings.Split(userInfos.LoadedPP, "/static/pp/")
		if len(parts) == 2 {
			oldFile := filepath.Join("static", "pp", parts[1])
			os.Remove(oldFile)
		}
	}
}

func deletePPHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if userInfos.Email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	removeOldPP(userInfos)
	userInfos.LoadedPP = ""
	userInfos.EditedPP = ""
	err := updateUserPP(userInfos.Email, "")
	if err != nil {
		userInfos.AccountError = "Erreur lors de la suppression de la photo."
	} else {
		userInfos.AccountError = "Photo de profil supprimée."
	}
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
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
