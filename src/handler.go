package forum

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func IsConnected(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	return cookie.Value != ""
}

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := struct {
		Page        string
		IsConnected bool
		Pseudo      string
	}{
		Page:        "home",
		IsConnected: IsConnected(r),
		Pseudo:      "Pseudo",
	}
	tmpl, err := template.ParseFiles("template/index.html", "template/header.html", "template/footer.html")
	if err != nil {
		http.Error(w, "Erreur template Home: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/login.html", "template/header.html", "template/footer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/register.html", "template/header.html", "template/footer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[len("/categories/"):]
	displayTitle := strings.ReplaceAll(slug, "-", " ")
	displayTitle = strings.Title(displayTitle)
	data := struct {
		Title       string
		Page        string
		IsConnected bool
		Pseudo      string
	}{
		Title:       displayTitle,
		Page:        "categories",
		IsConnected: IsConnected(r),
		Pseudo:      "Pseudo",
	}
	tmpl, _ := template.ParseFiles("template/category.html", "template/header.html", "template/footer.html")
	tmpl.ExecuteTemplate(w, "category.html", data)
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Page        string
		IsConnected bool
		Pseudo      string
	}{
		Page:        "forgot-password",
		IsConnected: IsConnected(r),
	}
	tmpl, _ := template.ParseFiles("template/forgot-pwd.html", "template/header.html", "template/footer.html")
	tmpl.Execute(w, data)
}

func SendResetLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forgot-pwd", http.StatusSeeOther)
		return
	}
	email := r.FormValue("email")

	fmt.Printf("\n[backend]: Demande de réinitialisation de mot de passe pour %s\n", email)

	token := "RESET-" + GenerateToken(email)

	fmt.Printf("[backend] Lien généré : http://localhost:8080/reset-pwd?token=%s\n\n", token)
	fmt.Fprint(w, "Si cet email existe, un lien a été envoyé.")

	/*println("EMAIL DE RECUPERATION")
	println("Destinataire :", email)
	println("Lien : http://localhost:8080/reset-pwd?token=" + token)
	println("------------------------------------")

	fmt.Fprint(w, "Un lien de récupération a été envoyé à votre adresse mail")*/
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
