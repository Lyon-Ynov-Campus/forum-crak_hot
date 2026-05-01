package forum

import (
	"html/template"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("index.html", "template/header.html", "template/footer.html")
	if err != nil {
		http.Error(w, "Erreur template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, "home")
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("login.html")
	tmpl.Execute(w, nil)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("register.html")
	tmpl.Execute(w, nil)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("category.html", "template/header.html", "template/footer.html")
	tmpl.Execute(w, nil)
}
