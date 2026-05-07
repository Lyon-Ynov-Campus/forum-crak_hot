package forum

import (
	"log"
	"net/http"
)

const serverAddr = "127.0.0.1:8080"

type UserInfos struct {
	Username              string
	EditedUsername        string
	Email                 string
	EditedEmail           string
	Email_Username        string
	Password              string
	EditedPassword        string
	ConfPassword          string
	ConfEditedPassword    string
	DeleteAccountPassword string
	LoadedPP              string
	EditedPP              string
	AccountError          string
	Status                string
	DBid                  string
	IsConnected           bool
}

var userInfos UserInfos

func StartServer() {
	http.HandleFunc("/api/like", API_LikeHandler)
	http.HandleFunc("/api/create-post", API_CreatePostHandler)
	http.HandleFunc("/api/posts", API_GetPostsHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { homeHandler(w, r, &userInfos) })
	http.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) { forumHandler(w, r, &userInfos) })
	http.HandleFunc("/reseau", NetworkHandler)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { loginHandler(w, r, &userInfos) })
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) { registerHandler(w, r, &userInfos) })
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) { logoutHandler(w, r, &userInfos) })
	http.HandleFunc("/postCreate", postCreate)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Server listening on http://%s", serverAddr)
	http.ListenAndServe(serverAddr, nil)
}
