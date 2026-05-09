package forum

import (
	"html/template"
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
	query := `SELECT username, email FROM Users INNER JOIN Session ON Users.id = Session.user_id WHERE Session.token = ?`
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
	db.QueryRow("SELECT id FROM Users WHERE email = ?", user.Email).Scan(&id)
	return id
}

func homeHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	user := GetUserFromSession(r)
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp
	tmpl, _ := template.ParseFiles("pages/index.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "index.html", struct {
		*UserInfos
		Page string
	}{user, "home"})
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	user := GetUserFromSession(r)
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp
	tmpl, _ := template.ParseFiles("pages/forum.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "forum.html", struct {
		*UserInfos
		Page string
	}{user, "home"})
}

func postCreate(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("pages/postCreate.html", "pages/header.html", "pages/footer.html")
		tmpl.ExecuteTemplate(w, "postCreate.html", struct {
			*UserInfos
			Page string
		}{user, "postCreate"})
		return
	}
	if r.Method == http.MethodPost {
		titre := r.FormValue("titre")
		contenu := r.FormValue("contenu")
		categorie := r.FormValue("categorie")
		date := time.Now().Format("2006-01-02")
		userID := GetUserID(r)
		CreatePost(titre, contenu, categorie, date, userID)
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
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
	tmpl, _ := template.ParseFiles("pages/post.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "post.html", struct {
		Post      Post
		Comments  []ComNameAuthor
		LikeCount int
		ComCount  int
		Author    string
		*UserInfos
		Page string
	}{post, comments, likeCount, comCount, authorPseudo, user, "post"})
}

func seeAllPosts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	posts, _ := GetAllPosts()
	var postsInfos []PostAllInfos
	for _, p := range posts {
		author, _ := GetPseudoByUserID(p.UserID)
		comCount, _ := CountCom(p.ID)
		likeCount, _ := CountLikes(p.ID)
		postsInfos = append(postsInfos, PostAllInfos{p.ID, p.Titre, p.Contenu, p.Categorie, p.DatePublication, author, comCount, likeCount})
	}
	tmpl, _ := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "category.html", struct {
		Posts []PostAllInfos
		*UserInfos
		Page string
	}{postsInfos, user, "posts"})
}

func likePost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	LikePost(userID, postID)
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func seeAllUsers(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	users, _ := GetAllUsers()
	tmpl, _ := template.ParseFiles("pages/seeAllUsers.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "seeAllUsers.html", struct {
		Users []User
		*UserInfos
		Page string
	}{users, user, "seeAllUsers"})
}

func seeUser(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	userID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	profile, _ := GetUserByID(userID)
	posts, _ := GetUserPosts(userID)
	comments, _ := GetUserCom(userID)
	tmpl, _ := template.ParseFiles("pages/seeUser.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "seeUser.html", struct {
		Profile  User
		Posts    []Post
		Comments []Com
		*UserInfos
		Page string
	}{profile, posts, comments, user, "seeUser"})
}

// Handlers Pages Spécifiques
func NetworkHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	tmpl, _ := template.ParseFiles("pages/reseau.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "reseau.html", struct {
		*UserInfos
		Page string
	}{user, "reseau"})
}

func HeartHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	// Ajoutez ici la logique SQL de HeartHandler pour hp (HeartPost)
	tmpl, _ := template.ParseFiles("pages/heart.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "heart.html", struct {
		*UserInfos
		Page string
	}{user, "heart"})
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	slug := r.URL.Path[len("/categories/"):]
	displayTitle := strings.Title(strings.ReplaceAll(slug, "-", " "))
	tmpl, _ := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	tmpl.ExecuteTemplate(w, "category.html", struct {
		Title string
		*UserInfos
		Page string
	}{displayTitle, user, "category"})
}

// Structures & Stubs
type ComNameAuthor struct {
	Contenu string
	DateCom string
	Author  string
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

func postUpdate(w http.ResponseWriter, r *http.Request) {}
func postDelete(w http.ResponseWriter, r *http.Request) {}
func myPosts(w http.ResponseWriter, r *http.Request)    {}
func comCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contenu := r.FormValue("contenu")
		postID, _ := strconv.Atoi(r.FormValue("post_id"))
		date := time.Now().Format("2006-01-02")
		userID := GetUserID(r)
		CreateCom(contenu, date, userID, postID)
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
	}
}
func comUpdate(w http.ResponseWriter, r *http.Request) {}
func comDelete(w http.ResponseWriter, r *http.Request) {}
func seeMyComs(w http.ResponseWriter, r *http.Request) {}
func unLikePost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	UnlikePost(userID, postID)
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}
func ForgotPasswordPage(w http.ResponseWriter, r *http.Request)   {}
func SendResetLink(w http.ResponseWriter, r *http.Request)        {}
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {}
