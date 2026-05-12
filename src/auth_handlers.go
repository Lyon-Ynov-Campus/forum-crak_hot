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
	LoadFlash(w, r, userInfos)
	tmpl, err := template.ParseFiles("pages/login.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur template login : %v", err)
		http.Error(w, "Erreur lors du chargement de la page", http.StatusInternalServerError)
		return
	}

	resetSent := r.URL.Query().Get("reset_sent") == "true"
	data := struct {
		*UserInfos
		IsConnected bool
		ResetSent   bool
		Query       string
	}{
		UserInfos:   userInfos,
		IsConnected: IsConnected(r),
		ResetSent:   resetSent,
		Query:       "",
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
	LoadFlash(w, r, userInfos)
	tmpl, err := template.ParseFiles("pages/register.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur template register : %v", err)
		return
	}
	data := struct {
		*UserInfos
		IsConnected bool
		Query       string
	}{
		UserInfos:   userInfos,
		IsConnected: IsConnected(r),
		Query:       "",
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
	userInfos.Username, userInfos.Email = "", ""
	userInfos.IsConnected = false
	userInfos.AccountError = ""

	SetFlash(w, "success", "Déconnecté avec succès.")

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
	if !user.IsConnected {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	LoadFlash(w, r, user)

	userID := GetUserID(r)
	myPosts, _ := GetUserPosts(userID)
	myComments, _ := GetUserCom(userID)

	pp, err := getUserPP(user.Email)
	user.LoadedPP = ""
	if err == nil {
		user.LoadedPP = pp
	}

	tmpl, err := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur template account : %v", err)
		return
	}

	data := struct {
		*UserInfos
		Page     string
		Posts    []Post
		Comments []Com
		Query    string
	}{
		UserInfos: user,
		Page:      "account",
		Posts:     myPosts,
		Comments:  myComments,
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "account.html", data)
}

func addPPHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	if userInfos.Email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		SetFlash(w, "error", "Erreur lors de l'upload.")
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	file, handler, err := r.FormFile("addPP")
	if err != nil {
		SetFlash(w, "error", "Erreur lors de la récupération du fichier.")
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		SetFlash(w, "error", "Format non supporté (Utilisez PNG ou JPG).")
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	ppDir := filepath.Join("static", "pp")
	os.MkdirAll(ppDir, 0755)

	filename := fmt.Sprintf("PPofNum%s%s", userInfos.DBid, ext)
	path := filepath.Join(ppDir, filename)

	removeOldPP(userInfos)

	out, err := os.Create(path)
	if err != nil {
		SetFlash(w, "error", "Erreur lors de la sauvegarde.")
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	ppURL := "/static/pp/" + filename
	updateUserPP(userInfos.Email, ppURL)

	SetFlash(w, "success", "Photo de profil mise à jour.")
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
	updateUserPP(userInfos.Email, "")
	SetFlash(w, "success", "Photo de profil supprimée.")
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func editusernameHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedUsername = r.FormValue("editedusername")
	dataEditUsername(w, r, userInfos)
	PushAccountFlash(w, userInfos)
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func editemailHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.EditedEmail = r.FormValue("editedemail")
	dataEditEmail(w, r, userInfos)
	PushAccountFlash(w, userInfos)
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func editpasswordHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}
	userInfos.EditedPassword = r.FormValue("editedpassword")
	userInfos.ConfEditedPassword = r.FormValue("confeditedpassword")
	dataEditPassword(w, r, userInfos)
	PushAccountFlash(w, userInfos)

	if userInfos.FlashType == "success" {
		logoutHandler(w, r, userInfos)
	} else {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
}

func deleteaccountHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	userInfos.DeleteAccountPassword = r.FormValue("deleteaccountpassword")
	dataDeleteAccount(w, r, userInfos)
	PushAccountFlash(w, userInfos)
	if userInfos.FlashType == "success" {
		logoutHandler(w, r, userInfos)
	} else {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
	}
}
