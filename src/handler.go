package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func IsConnected(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	return true
}

func homeHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("pages/index.html", "pages/header.html", "pages/footer.html")
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

	tmpl.ExecuteTemplate(w, "index.html", data)
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("pages/forum.html", "pages/header.html", "pages/footer.html")
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

	tmpl.ExecuteTemplate(w, "forum.html", data)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[len("/categories/"):]
	displayTitle := strings.ReplaceAll(slug, "-", " ")
	displayTitle = strings.Title(displayTitle)

	pseudo := "Invité"
	if IsConnected(r) {
		cookie, _ := r.Cookie("session_token")
		decoded, _ := ValidateToken(cookie.Value)
		pseudo = strings.Split(decoded, "|")[0]
	}

	data := struct {
		Title       string
		IsConnected bool
		Username    string
	}{
		Title:       displayTitle,
		IsConnected: IsConnected(r),
		Username:    pseudo,
	}

	tmpl, _ := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "category.html", data)
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Page        string
		IsConnected bool
	}{
		Page:        "forgot-password",
		IsConnected: IsConnected(r),
	}
	tmpl, _ := template.ParseFiles("pages/forgot-pwd.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "forgot-pwd.html", data)
}

func SendResetLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}
	email := r.FormValue("email")

	fmt.Printf("\n[backend]: Demande de réinitialisation pour %s\n", email)
	token := "RESET-" + GenerateToken(email)
	fmt.Printf("[backend] Lien généré : http://localhost:8080/reset-pwd?token=%s\n\n", token)

	fmt.Fprint(w, "Si cet email existe, un lien a été envoyé dans votre terminal.")
}

func FakeLogin(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "token-test",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
