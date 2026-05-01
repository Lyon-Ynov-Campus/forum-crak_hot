package forum

import (
	"html/template"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("./pages/index.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, userInfos)
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	tmpl, err := template.ParseFiles("./pages/forum.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, userInfos)
}
