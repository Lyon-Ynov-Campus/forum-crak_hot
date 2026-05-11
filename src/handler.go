package forum

import (
	"html/template"
	"log"
	"net/http"
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
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp
	tmpl, err := template.ParseFiles("pages/index.html", "pages/header.html", "pages/footer.html")
	if err != nil { log.Fatal(err) }
	tmpl.ExecuteTemplate(w, "index.html", struct {
		*UserInfos
		Page string
	}{user, "home"})
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	posts, _ := GetAllPosts()
	user := GetUserFromSession(r)
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
	tmpl, _ := template.ParseFiles("pages/forum.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "forum.html", struct {
		Posts []PostAllInfos; *UserInfos; Page string
	}{postsAllInfos, user, "forum"})
}

func HeartHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	type HeartPost struct {
		ID int; Titre, Contenu string; LikeCount int; Date, Auteur string
	}
	var hp HeartPost
	query := `
    SELECT p.id, p.titre, p.contenu, p.date_publication, u.username, COUNT(l.id) as total_likes
    FROM Post p
    LEFT JOIN Like l ON p.id = l.post_id
    LEFT JOIN Users u ON p.user_id = u.id
    GROUP BY p.id
    ORDER BY total_likes DESC, p.date_publication DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&hp.ID, &hp.Titre, &hp.Contenu, &hp.Date, &hp.Auteur, &hp.LikeCount)
	if err != nil {
		hp = HeartPost{Titre: "Pas encore de favori", Contenu: "Faites vivre le forum pour voir apparaître un coup de cœur !"}
	}
	tmpl, _ := template.ParseFiles("pages/heart.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "heart.html", struct { *UserInfos; Page string; Post HeartPost }{user, "heart", hp})
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	resetSent := r.URL.Query().Get("reset_sent") == "true"
	tmpl, _ := template.ParseFiles("pages/forgot-pwd.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "forgot-pwd.html", struct { *UserInfos; Page string; ResetSent bool }{user, "forgot-password", resetSent})
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		user := GetUserFromSession(r)
		token := r.URL.Query().Get("token")
		_, isValid := ValidatePasswordResetToken(token)
		tmpl, _ := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
		tmpl.ExecuteTemplate(w, "reset-password.html", struct { *UserInfos; Token, ResetError, ResetSuccess string }{
			user, token, func()string{if !isValid{return "Lien invalide ou expiré"} ; return ""}(), "",
		})
	} else if r.Method == http.MethodPost {
		token, pass, conf := r.FormValue("token"), r.FormValue("password"), r.FormValue("confirm_password")
		email, isValid := ValidatePasswordResetToken(token)
		if !isValid { http.Redirect(w, r, "/forgot-password", 303); return }
		err := ResetPassword(email, pass, conf)
		tmpl, _ := template.ParseFiles("pages/reset-password.html", "pages/header.html", "pages/footer.html")
		tmpl.ExecuteTemplate(w, "reset-password.html", struct { *UserInfos; Token, ResetError, ResetSuccess string }{
			&UserInfos{}, token, err, func()string{if err==""{return "Réinitialisé avec succès !"};return ""}(),
		})
	}
}

func seeOnePost(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
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
	if userID := GetUserID(r); userID != 0 { liked = HasLiked(userID, postID) }
	tmpl, _ := template.ParseFiles("pages/post.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "post.html", struct {
		Post Post; Comments []ComNameAuthor; LikeCount, ComCount int; Author string; Liked bool; *UserInfos; Page string
	}{post, comments, likeCount, comCount, authorPseudo, liked, user, "post"})
}

func myPosts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	posts, _ := GetUserPosts(GetUserID(r))
	tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "account.html", struct { Posts []Post; *UserInfos; Page string }{posts, user, "myPosts"})
}

type ComNameAuthor struct { Contenu, DateCom, Author string }
type PostAllInfos struct { ID int; Titre, Contenu, Categorie, DatePublication, Author string; ComCount, LikeCount int }

func postCreate(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("pages/postCreate.html", "pages/header.html", "pages/footer.html")
		tmpl.ExecuteTemplate(w, "postCreate.html", struct { *UserInfos; Page string }{user, "postCreate"})
	} else if r.Method == http.MethodPost {
		CreatePost(r.FormValue("titre"), r.FormValue("contenu"), r.FormValue("categorie"), time.Now().Format("2006-01-02"), GetUserID(r))
		http.Redirect(w, r, "/posts", 303)
	}
}

func postUpdate(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.FormValue("id"))
	post, _ := GetPostByID(postID)
	if post.UserID != GetUserID(r) { http.Redirect(w, r, "/myPosts", 303); return }
	UpdatePost(postID, r.FormValue("titre"), r.FormValue("contenu"), r.FormValue("categorie"))
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), 303)
}

func postDelete(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	post, _ := GetPostByID(postID)
	if post.UserID == GetUserID(r) { DeletePost(postID) }
	http.Redirect(w, r, "/myPosts", 303)
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

func SendResetLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		DataForgotPasswordSend(w, r, email)
		http.Redirect(w, r, "/login?reset_sent=true", http.StatusSeeOther)
	}
}

func seeAllPosts(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func comCreate(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func comUpdate(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func comDelete(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func seeMyComs(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func seeUser(w http.ResponseWriter, r *http.Request) {}
func seeAllUsers(w http.ResponseWriter, r *http.Request) {}
func likePost(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("id")
    http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}

func unLikePost(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("id")
    http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}