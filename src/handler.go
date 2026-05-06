package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func GetUserFromSession(r *http.Request) *UserInfos {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return &UserInfos{Username: "Invité", IsConnected: false}
	}

	var user UserInfos
	query := `
        SELECT Users.username, Users.email 
        FROM Users 
        INNER JOIN Session ON Users.id = Session.user_id 
        WHERE Session.token = ?`

	err = db.QueryRow(query, cookie.Value).Scan(&user.Username, &user.Email)
	userInfos.Username = user.Username
	userInfos.Email = user.Email

	if err != nil {
		fmt.Println("Debug Auth:", err)
		return &UserInfos{Username: "Invité", IsConnected: false}
	}

	user.IsConnected = true
	return &user
}

func IsConnected(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	return true
}

func homeHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)

	pp, err := getUserPP(user.Email)
	if err == nil {
		user.LoadedPP = pp
	} else {
		user.LoadedPP = ""
	}

	tmpl, err := template.ParseFiles("pages/index.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		*UserInfos
		Page string
	}{
		UserInfos: user,
		Page:      "home",
	}
	tmpl.ExecuteTemplate(w, "index.html", data)
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)

	pp, err := getUserPP(user.Email)
	if err == nil {
		user.LoadedPP = pp
	} else {
		user.LoadedPP = ""
	}

	tmpl, err := template.ParseFiles("pages/forum.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		*UserInfos
		Page string
	}{
		UserInfos: user,
		Page:      "home",
	}

	tmpl.ExecuteTemplate(w, "forum.html", data)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)

	slug := r.URL.Path[len("/categories/"):]
	displayTitle := strings.ReplaceAll(slug, "-", " ")
	displayTitle = strings.Title(displayTitle)

	data := struct {
		Title string
		*UserInfos
		Page string
	}{
		Title:     displayTitle,
		UserInfos: user,
		Page:      "category", // Aucune bouton du menu principal ne sera en dégradé
	}

	tmpl, err := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		fmt.Println("Erreur template:", err)
		return
	}
	tmpl.ExecuteTemplate(w, "category.html", data)
}

func NetworkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)

	data := struct {
		*UserInfos
		Page string
	}{
		UserInfos: user,
		Page:      "reseau",
	}

	tmpl, err := template.ParseFiles("pages/reseau.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		fmt.Println("Erreur template:", err)
		return
	}
	tmpl.ExecuteTemplate(w, "reseau.html", data)
}

func HeartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)

	type HeartPost struct {
		ID        int
		Titre     string
		Contenu   string
		LikeCount int
		Date      string
		Auteur    string
	}

	var hp HeartPost

	query := `
    SELECT p.id, p.titre, p.contenu, p.date_publication, u.username, COUNT(l.id) as total_likes
    FROM Post p
    LEFT JOIN Like l ON p.id = l.post_id
    LEFT JOIN Users u ON p.user_id = u.id
    GROUP BY p.id
    ORDER BY total_likes DESC, p.date_publication DESC
    LIMIT 1`

	err := db.QueryRow(query).Scan(&hp.ID, &hp.Titre, &hp.Contenu, &hp.Date, &hp.Auteur, &hp.LikeCount)

	if err != nil {
		hp = HeartPost{Titre: "Pas encore de favori", Contenu: "Faites vivre le forum pour voir apparaître un coup de cœur !"}
	}

	tmpl, err := template.ParseFiles("pages/heart.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		fmt.Println("Erreur template:", err)
		return
	}

	data := struct {
		*UserInfos
		Page string
		Post HeartPost
	}{
		UserInfos: user,
		Page:      "heart",
		Post:      hp,
	}

	tmpl.ExecuteTemplate(w, "heart.html", data)
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)

	resetSent := r.URL.Query().Get("reset_sent") == "true"

	data := struct {
		*UserInfos
		Page      string
		ResetSent bool
	}{
		UserInfos: user,
		Page:      "forgot-password",
		ResetSent: resetSent,
	}

	tmpl, err := template.ParseFiles("pages/forgot-pwd.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		fmt.Println("Erreur template:", err)
		return
	}
	tmpl.ExecuteTemplate(w, "forgot-pwd.html", data)
}

func SendResetLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}
	email := r.FormValue("email")

	DataForgotPasswordSend(w, r, email)

	http.Redirect(w, r, "/login?reset_sent=true", http.StatusSeeOther)
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

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodGet {
		user := GetUserFromSession(r)

		token := r.URL.Query().Get("token")

		_, isValid := ValidatePasswordResetToken(token)

		if !isValid {
			data := struct {
				*UserInfos
				ResetError string
			}{
				UserInfos:  user,
				ResetError: "Lien de réinitialisation invalide ou expiré.",
			}

			tmpl, err := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
			if err != nil {
				fmt.Println("Erreur template:", err)
				return
			}
			tmpl.ExecuteTemplate(w, "reset-password.html", data)
			return
		}

		data := struct {
			*UserInfos
			Token        string
			ResetError   string
			ResetSuccess string
		}{
			UserInfos:    user,
			Token:        token,
			ResetError:   "",
			ResetSuccess: "",
		}

		tmpl, err := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
		if err != nil {
			fmt.Println("Erreur template:", err)
			return
		}
		tmpl.ExecuteTemplate(w, "reset-password.html", data)
		return
	}

	if r.Method == http.MethodPost {
		token := r.FormValue("token")
		newPassword := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		email, isValid := ValidatePasswordResetToken(token)

		if !isValid {
			data := struct {
				*UserInfos
				ResetError string
			}{
				UserInfos:  &UserInfos{},
				ResetError: "Lien de réinitialisation invalide ou expiré.",
			}

			tmpl, err := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
			if err != nil {
				fmt.Println("Erreur template:", err)
				return
			}
			tmpl.ExecuteTemplate(w, "reset-password.html", data)
			return
		}

		errorMsg := ResetPassword(email, newPassword, confirmPassword)

		if errorMsg != "" {
			data := struct {
				*UserInfos
				Token        string
				ResetError   string
				ResetSuccess string
			}{
				UserInfos:    &UserInfos{},
				Token:        token,
				ResetError:   errorMsg,
				ResetSuccess: "",
			}

			tmpl, err := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
			if err != nil {
				fmt.Println("Erreur template:", err)
				return
			}
			tmpl.ExecuteTemplate(w, "reset-password.html", data)
			return
		}

		data := struct {
			*UserInfos
			ResetSuccess string
		}{
			UserInfos:    &UserInfos{},
			ResetSuccess: "Votre mot de passe a été réinitialisé avec succès ! Vous pouvez maintenant vous connecter.",
		}

		tmpl, err := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
		if err != nil {
			fmt.Println("Erreur template:", err)
			return
		}
		tmpl.ExecuteTemplate(w, "reset-password.html", data)
		return
	}

	http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
}
