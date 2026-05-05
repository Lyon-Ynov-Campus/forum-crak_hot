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
	AccountError          string
	Status                string
	DBid                  string
	IsConnected           bool
}

var userInfos UserInfos

func StartServer() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homeHandler(w, r, &userInfos)
	})
	http.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) {
		forumHandler(w, r, &userInfos)
	})
	http.HandleFunc("/categories/", CategoryHandler)
	http.HandleFunc("/forgot-password", ForgotPasswordPage)
	http.HandleFunc("/send-reset", SendResetLink)
	http.HandleFunc("/reseau", NetworkHandler)
	http.HandleFunc("/coup-de-coeur", HeartHandler)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, &userInfos)
	})
	http.HandleFunc("/checklogin", func(w http.ResponseWriter, r *http.Request) {
		checkloginHandler(w, r, &userInfos)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		registerHandler(w, r, &userInfos)
	})
	http.HandleFunc("/checkregister", func(w http.ResponseWriter, r *http.Request) {
		checkregisterHandler(w, r, &userInfos)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(w, r, &userInfos)
	})

	http.HandleFunc("/editaccount", func(w http.ResponseWriter, r *http.Request) {
		editaccountHandler(w, r, &userInfos)
	})
	http.HandleFunc("/editusername", func(w http.ResponseWriter, r *http.Request) {
		editusernameHandler(w, r, &userInfos)
	})
	http.HandleFunc("/editemail", func(w http.ResponseWriter, r *http.Request) {
		editemailHandler(w, r, &userInfos)
	})
	http.HandleFunc("/editpassword", func(w http.ResponseWriter, r *http.Request) {
		editpasswordHandler(w, r, &userInfos)
	})
	http.HandleFunc("/deleteaccount", func(w http.ResponseWriter, r *http.Request) {
		deleteaccountHandler(w, r, &userInfos)
	})

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	as := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", as))

	log.Printf("Server listening on http://%s", serverAddr)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
