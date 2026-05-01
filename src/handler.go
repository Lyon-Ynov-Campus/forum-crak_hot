package forum

import (
	"html/template"
	"net/http"
)

func render(w http.ResponseWriter, filename string) {
	tmpl, err := template.ParseFiles(filename, "template/header.html", "template/footer.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement des templates:"+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	render(w, "index.html")
	// tmpl, err := template.ParseFiles("index.html")

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// tmpl.Execute(w, nil)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	render(w, "login.html")
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	render(w, "register.html")
}
