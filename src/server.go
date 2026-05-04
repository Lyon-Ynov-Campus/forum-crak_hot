package forum

import (
	"fmt"
	"net/http"
)

func Server() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/login", LoginPage)
	http.HandleFunc("/register", RegisterPage)
	http.HandleFunc("/categories/", CategoryHandler)

	http.HandleFunc("/fake-login", FakeLogin)
	http.HandleFunc("/fake-logout", Logout)

	as := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", as))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Le serveur est lancé sur http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Erreur serveur: %v\n", err)
	}
}
