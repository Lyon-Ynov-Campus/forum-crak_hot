package forum

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func IsConnected(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	return true
}

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
	if err != nil {
		return &UserInfos{Username: "Invité", IsConnected: false}
	}

	user.IsConnected = true
	return &user
}

func GetUserID(r *http.Request) int {
	user := GetUserFromSession(r)
	if !user.IsConnected {
		return 0
	}
	var id int
	err := db.QueryRow("SELECT id FROM Users WHERE email = ?", user.Email).Scan(&id)
	if err != nil {
		return 0
	}
	return id
}

func homeHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	tmpl, err := template.ParseFiles("pages/index.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Fatal(err)
	}
	tmpl.ExecuteTemplate(w, "index.html", struct {
		*UserInfos
		Page string
	}{user, "home"})
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)
	posts, _ := GetAllPosts()
	
	var postsAllInfos []PostAllInfos
	for _, p := range posts {
		pseudo, _ := GetPseudoByUserID(p.UserID)
		comCount, _ := CountCom(p.ID)
		likeCount, _ := CountLikes(p.ID)
		postsAllInfos = append(postsAllInfos, PostAllInfos{
			ID: p.ID, Titre: p.Titre, Author: pseudo, Categorie: p.Categorie,
			ComCount: comCount, LikeCount: likeCount, DatePublication: p.DatePublication,
		})
	}

	tmpl, err := template.ParseFiles("pages/forum.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Fatal(err)
	}
	tmpl.ExecuteTemplate(w, "forum.html", struct {
		Posts []PostAllInfos
		*UserInfos
		Page string
	}{postsAllInfos, user, "forum"})
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)
	resetSent := r.URL.Query().Get("reset_sent") == "true"

	tmpl, _ := template.ParseFiles("pages/forgot-pwd.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "forgot-pwd.html", struct {
		*UserInfos
		Page      string
		ResetSent bool
	}{user, "forgot-password", resetSent})
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		user := GetUserFromSession(r)
		LoadFlash(w, r, user)
		token := r.URL.Query().Get("token")
		_, isValid := ValidatePasswordResetToken(token)

		tmpl, _ := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
		tmpl.ExecuteTemplate(w, "reset-password.html", struct {
			*UserInfos
			Token        string
			ResetError   string
			ResetSuccess string
		}{user, token, func() string {
			if !isValid { return "Lien invalide ou expiré" }
			return ""
		}(), ""})
		return
	}

	if r.Method == http.MethodPost {
		token := r.FormValue("token")
		pass, conf := r.FormValue("password"), r.FormValue("confirm_password")
		email, isValid := ValidatePasswordResetToken(token)
		
		if !isValid {
			SetFlash(w, "error", "Lien invalide ou expiré.")
			http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
			return
		}

		err := ResetPassword(email, pass, conf)
		if err != "" {
			SetFlash(w, "error", err)
			http.Redirect(w, r, "/reset-password?token="+url.QueryEscape(token), http.StatusSeeOther)
			return
		}

		SetFlash(w, "success", "Mot de passe réinitialisé avec succès !")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func seeOnePost(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)
	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	post, _ := GetPostByID(postID)
	
	rawComments, _ := GetComByPostID(postID)
	var comments []ComNameAuthor
	for _, c := range rawComments {
		pseudo, _ := GetPseudoByUserID(c.UserID)
		comments = append(comments, ComNameAuthor{c.Contenu, c.DateCom, pseudo})
	}

	likeCount, _ := CountLikes(postID)
	comCount, _ := CountCom(postID)
	authorPseudo, _ := GetPseudoByUserID(post.UserID)
	
	liked := false
	if userID := GetUserID(r); userID != 0 {
		liked = HasLiked(userID, postID)
	}

	tmpl, _ := template.ParseFiles("pages/post.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "post.html", struct {
		Post      Post
		Comments  []ComNameAuthor
		LikeCount int
		ComCount  int
		Author    string
		Liked     bool
		*UserInfos
		Page string
	}{post, comments, likeCount, comCount, authorPseudo, liked, user, "post"})
}

func postCreate(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("pages/postCreate.html", "pages/header.html", "pages/footer.html")
		tmpl.ExecuteTemplate(w, "postCreate.html", struct {
			*UserInfos
			Page string
		}{user, "postCreate"})
		return
	}

	if r.Method == http.MethodPost {
		CreatePost(r.FormValue("titre"), r.FormValue("contenu"), r.FormValue("categorie"), time.Now().Format("2006-01-02"), GetUserID(r))
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	}
}

type ComNameAuthor struct {
	Contenu, DateCom, Author string
}

type PostAllInfos struct {
	ID              int
	Titre           string
	Contenu         string
	Categorie       string
	DatePublication string
	Author          string
	ComCount        int
	LikeCount       int
}

func postUpdate(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/forum", 303) }
func postDelete(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/forum", 303) }
func myPosts(w http.ResponseWriter, r *http.Request)    { http.Redirect(w, r, "/forum", 303) }

func seeAllPosts(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/forum", 303) }
func comCreate(w http.ResponseWriter, r *http.Request)    { http.Redirect(w, r, "/forum", 303) }
func comUpdate(w http.ResponseWriter, r *http.Request)    { http.Redirect(w, r, "/forum", 303) }
func comDelete(w http.ResponseWriter, r *http.Request)    { http.Redirect(w, r, "/forum", 303) }
func seeMyComs(w http.ResponseWriter, r *http.Request)    { http.Redirect(w, r, "/forum", 303) }
func likePost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	http.Redirect(w, r, "/post?id="+id, 303)
}
func unLikePost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	http.Redirect(w, r, "/post?id="+id, 303)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	slug := r.URL.Path[len("/categories/"):]
	displayTitle := strings.Title(strings.ReplaceAll(slug, "-", " "))
	tmpl, _ := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "category.html", struct {
		Title string; *UserInfos; Page string
	}{displayTitle, user, "category"})
}

func NetworkHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	tmpl, _ := template.ParseFiles("pages/reseau.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "reseau.html", struct { *UserInfos; Page string }{user, "reseau"})
}

func HeartHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	// Logique simplifiée pour l'exemple
	tmpl, _ := template.ParseFiles("pages/heart.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "heart.html", struct { *UserInfos; Page string }{user, "heart"})
}

func SendResetLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		DataForgotPasswordSend(w, r, email)
		http.Redirect(w, r, "/login?reset_sent=true", http.StatusSeeOther)
	}
}

func seeUser(w http.ResponseWriter, r *http.Request)     { http.Redirect(w, r, "/reseau", 303) }
func seeAllUsers(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/reseau", 303) }