package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

// Gestion action user
//partie post

func GetUserID(r *http.Request) int { //func faite par IA aide a recup ID du user dans BDD pr chaque etape des handlers
	user := GetUserFromSession(r)
	if !user.IsConnected {
		return 0
	}

	var id int
	err := db.QueryRow("SELECT id FROM Users WHERE email = ?", user.Email).Scan(&id)
	if err != nil {
		fmt.Println("err GetUserID", err)
		return 0
	}
	return id
}

func postCreate(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r) //recup info user

	if r.Method == http.MethodGet { //affiche forualire si arrvie sur la page
		tmpl, _ := template.ParseFiles("pages/postCreate.html", "pages/header.html", "pages/footer.html")
		data := struct {
			*UserInfos
			Page string
		}{
			UserInfos: user,         //affiche pseudo dans header
			Page:      "postCreate", //active bon btn dans la nav
		}
		tmpl.ExecuteTemplate(w, "postCreate.html", data)
		return
	}

	if r.Method == http.MethodPost { //si form envoyer
		titre := r.FormValue("titre") //recup titre etc
		contenu := r.FormValue("contenu")
		categorie := r.FormValue("categorie")
		date := time.Now().Format("2006-01-02") //rev src codystudy.net

		userID := GetUserID(r) //recup id du user

		CreatePost(titre, contenu, categorie, date, userID) //créer le post

		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	}
}

func postUpdate(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	postID, _ := strconv.Atoi(r.URL.Query().Get("id")) //recup id du post ds URL
	post, err := GetPostByID(postID)                   //recup post contenu dnas BDD

	if err != nil { //si psot existe pas
		http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
		return
	}

	if post.UserID != GetUserID(r) { //verif si bien auteeur car server public
		http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
		data := struct {
			Post Post
			*UserInfos
			Page string
		}{
			Post:      post, //pre remplir le form
			UserInfos: user, //info user pour header car nom a coté de pp revoir figma
			Page:      "postUpdate",
		}
		tmpl.ExecuteTemplate(w, "account.html", data)
		return
	}

	if r.Method == http.MethodPost {
		newTitre := r.FormValue("titre")
		newContenu := r.FormValue("contenu")
		newCategorie := r.FormValue("categorie")

		UpdatePost(postID, newTitre, newContenu, newCategorie)

		http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
	}
}

func postDelete(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	post, err := GetPostByID(postID)
	if err != nil {
		http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
		return
	}

	// seule sécurité nécessaire : vérifier que c’est l’auteur
	if post.UserID != GetUserID(r) { //aide IA ms ps forcmetn necessaire a rev
		http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
		return
	}

	DeletePost(postID)
	http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
}

func seeOnePost(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	post, err := GetPostByID(postID)
	if err != nil {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
		return
	}

	rawComments, _ := GetComByPostID(postID) //ajout partie recup commetnaires

	var comments []ComNameAuthor
	for _, c := range rawComments {
		pseudo, _ := GetPseudoByUserID(c.UserID)

		comments = append(comments, ComNameAuthor{
			Contenu: c.Contenu,
			DateCom: c.DateCom,
			Author:  pseudo,
		})
	}

	likeCount, _ := CountLikes(postID)
	comCount, _ := CountCom(postID) //peut etre pas necessaire a voir pr enelver apres
	authorPseudo, _ := GetPseudoByUserID(post.UserID)

	tmpl, _ := template.ParseFiles("pages/post.html", "pages/header.html", "pages/footer.html")
	data := struct {
		Post      Post
		Comments  []ComNameAuthor
		LikeCount int
		ComCount  int
		Author    string
		*UserInfos
		Page string
	}{
		Post:      post,
		Comments:  comments,
		LikeCount: likeCount,
		ComCount:  comCount,
		Author:    authorPseudo,
		UserInfos: user,
		Page:      "post",
	}

	tmpl.ExecuteTemplate(w, "post.html", data)
}

type ComNameAuthor struct { //rajout comme post ci dessus pour avoir nom du user
	Contenu string
	DateCom string
	Author  string
}

type PostAllInfos struct { //rajout pr tout info de affiche d'un post pr avoir le nom du user
	ID              int
	Titre           string
	Contenu         string
	Categorie       string
	DatePublication string
	Author          string
	ComCount        int
	LikeCount       int
}

func seeAllPosts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	posts, _ := GetAllPosts()

	var postsInfos []PostAllInfos

	for _, p := range posts {
		author, _ := GetPseudoByUserID(p.UserID)
		comCount, _ := CountCom(p.ID)
		likeCount, _ := CountLikes(p.ID)

		postsInfos = append(postsInfos, PostAllInfos{
			ID:              p.ID,
			Titre:           p.Titre,
			Contenu:         p.Contenu,
			Categorie:       p.Categorie,
			DatePublication: p.DatePublication,
			Author:          author,
			ComCount:        comCount,
			LikeCount:       likeCount,
		})
	}

	tmpl, _ := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	data := struct {
		Posts []PostAllInfos
		*UserInfos
		Page string
	}{
		Posts:     postsInfos,
		UserInfos: user,
		Page:      "posts",
	}

	tmpl.ExecuteTemplate(w, "category.html", data)
}

func myPosts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r) //recup user

	posts, _ := GetUserPosts(GetUserID(r)) //recuperer SES posts

	tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
	data := struct {
		Posts []Post
		*UserInfos
		Page string
	}{
		Posts:     posts, //ses posts
		UserInfos: user,  //info user
		Page:      "myPosts",
	}

	tmpl.ExecuteTemplate(w, "account.html", data)
}

//partie com reper

func comCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contenu := r.FormValue("contenu")
		postID, _ := strconv.Atoi(r.FormValue("post_id"))
		date := time.Now().Format("2006-01-02")

		userID := GetUserID(r)

		CreateCom(contenu, date, userID, postID)

		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther) //secu si qlq accede en get car get pas id pas comme post donc au cas ou inspriation IA
}

func comUpdate(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	comID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	com, err := GetComByID(comID)

	if err != nil || com.UserID != GetUserID(r) { //verif que c auteru come post
		http.Redirect(w, r, "/myComs", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
		data := struct {
			Com Com
			*UserInfos
			Page string
		}{
			Com:       com,
			UserInfos: user,
			Page:      "comUpdate",
		}
		tmpl.ExecuteTemplate(w, "account.html", data)
		return
	}

	if r.Method == http.MethodPost {
		newContenu := r.FormValue("contenu")

		UpdateCom(comID, newContenu)

		http.Redirect(w, r, "/myComs", http.StatusSeeOther)
	}
}

func comDelete(w http.ResponseWriter, r *http.Request) {
	comID, _ := strconv.Atoi(r.URL.Query().Get("id")) //meme log que pr post
	com, err := GetComByID(comID)

	if err != nil {
		http.Redirect(w, r, "/myComs", http.StatusSeeOther)
		return
	}

	if com.UserID != GetUserID(r) { // vérifie que c'est l'auteur
		http.Redirect(w, r, "/myComs", http.StatusSeeOther)
		return
	}

	DeleteCom(comID)
	http.Redirect(w, r, "/myComs", http.StatusSeeOther) //reviens tjs a chaque fois pour voir si bien supprimer
}

func seeMyComs(w http.ResponseWriter, r *http.Request) { //page profil user
	user := GetUserFromSession(r)

	comments, _ := GetUserCom(GetUserID(r))

	tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
	data := struct {
		Comments []Com
		*UserInfos
		Page string
	}{
		Comments:  comments,
		UserInfos: user,
		Page:      "myComs",
	}

	tmpl.ExecuteTemplate(w, "account.html", data)
}

//partie like

func likePost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))

	LikePost(userID, postID)

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func unLikePost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))

	UnlikePost(userID, postID)

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

//partie recherche réseau

func seeAllUsers(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	users, _ := GetAllUsers()

	tmpl, _ := template.ParseFiles("pages/seeAllUsers.html", "pages/header.html", "pages/footer.html")
	data := struct {
		Users []User
		*UserInfos
		Page string
	}{
		Users:     users,
		UserInfos: user,
		Page:      "seeAllUsers",
	}

	tmpl.ExecuteTemplate(w, "seeAllUsers.html", data)
}

func seeUser(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	userID, _ := strconv.Atoi(r.URL.Query().Get("id"))

	profile, err := GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/seeAllUsers", http.StatusSeeOther)
		return
	}

	posts, _ := GetUserPosts(userID)
	comments, _ := GetUserCom(userID)

	tmpl, _ := template.ParseFiles("pages/seeUser.html", "pages/header.html", "pages/footer.html")
	data := struct {
		Profile  User
		Posts    []Post
		Comments []Com
		*UserInfos
		Page string
	}{
		Profile:   profile,
		Posts:     posts,
		Comments:  comments,
		UserInfos: user,
		Page:      "seeUser",
	}

	tmpl.ExecuteTemplate(w, "seeUser.html", data)
}
