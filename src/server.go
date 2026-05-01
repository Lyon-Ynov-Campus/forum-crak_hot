package forum

import (
	"fmt"
	"net/http"
)

func Server() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/login", LoginPage)
	http.HandleFunc("/register", RegisterPage)
	http.HandleFunc("/", Home)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Serveur lancé sur localhost 8080")
	http.ListenAndServe(":8080", nil)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur serveur: %v\n", err)
	}
}
