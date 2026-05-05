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

//action user partie

func postCreate(w http.ResponseWriter, r *http.Request) {

}

func postUpdate(w http.ResponseWriter, r *http.Request) {

}

func postDelete(w http.ResponseWriter, r *http.Request) {

}

func myPosts(w http.ResponseWriter, r *http.Request) {

}

func posts(w http.ResponseWriter, r *http.Request) {

}

func seeOnePost(w http.ResponseWriter, r *http.Request) {

}

func seeAllPosts(w http.ResponseWriter, r *http.Request) {

}

func comCreate(w http.ResponseWriter, r *http.Request) {

}

func comUpdate(w http.ResponseWriter, r *http.Request) {

}

func comDelete(w http.ResponseWriter, r *http.Request) {

}

func Like(w http.ResponseWriter, r *http.Request) {

}

func unLike(w http.ResponseWriter, r *http.Request) {

}

func seeUser(w http.ResponseWriter, r *http.Request) {

}

func seeAllUsers(w http.ResponseWriter, r *http.Request) {

}
