package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	LoadFlash(w, r, user)

	pp, err := getUserPP(user.Email)
	if err == nil {
		user.LoadedPP = pp
	} else {
		user.LoadedPP = ""
	}
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func forumHandler(w http.ResponseWriter, r *http.Request, userInfos *UserInfos) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	search := r.URL.Query().Get("q")  //ajout pr tri
	sort := r.URL.Query().Get("sort") //ajout pr tri
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)

	pp, err := getUserPP(user.Email)
	if err == nil {
		user.LoadedPP = pp
	} else {
		user.LoadedPP = ""
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))

	var posts []Post

	if query == "" {
		posts, _ = GetAllPosts()
	} else {
		posts, _ = SearchPostsByTitle(query)
	}

	for i := range posts { //rempli val avant le tri car sinon func getallposts et SearchPostsByTitle ne remplisse pas count like ou com
		posts[i].CountCom, _ = CountCom(posts[i].ID)
		posts[i].CountLikes, _ = CountLikes(posts[i].ID)
	}

	posts = ApplyTri(posts, sort) //ajout pr tri

	var postsAllInfos []PostAllInfos

	for _, p := range posts {
		pseudo, _ := GetPseudoByUserID(p.UserID)
		comCount, _ := CountCom(p.ID)
		likeCount, _ := CountLikes(p.ID)

		postsAllInfos = append(postsAllInfos, PostAllInfos{
			ID:              p.ID,
			Titre:           p.Titre,
			Author:          pseudo,
			Categorie:       p.Categorie,
			ComCount:        comCount,
			LikeCount:       likeCount,
			DatePublication: p.DatePublication,
		})
	}

	tmpl, _ := template.ParseFiles("pages/forum.html", "pages/header.html", "pages/footer.html")

	data := struct {
		Posts  []PostAllInfos
		Search string
		Sort   string
		*UserInfos
		Page  string
		Query string
	}{
		Posts:     postsAllInfos,
		Search:    search,
		Sort:      sort,
		UserInfos: user,
		Page:      "home",
		Query:     query,
	}

	tmpl.ExecuteTemplate(w, "forum.html", data)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp
	LoadFlash(w, r, user)

	slug := r.URL.Path[len("/categories/"):]
	displayTitle := strings.ReplaceAll(slug, "-", " ")
	displayTitle = strings.Title(displayTitle)

	data := struct {
		Title string
		*UserInfos
		Page  string
		Query string
	}{
		Title:     displayTitle,
		UserInfos: user,
		Page:      "category",
		Query:     "",
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

	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	LoadFlash(w, r, user)

	data := struct {
		*UserInfos
		Page  string
		Query string
	}{
		UserInfos: user,
		Page:      "reseau",
		Query:     "",
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

	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	LoadFlash(w, r, user)

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
		Page  string
		Query string
	}{
		UserInfos: user,
		Page:      "heart",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "heart.html", data)
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := GetUserFromSession(r)
	LoadFlash(w, r, user)

	resetSent := r.URL.Query().Get("reset_sent") == "true"

	data := struct {
		*UserInfos
		Page      string
		ResetSent bool
		Query     string
	}{
		UserInfos: user,
		Page:      "forgot-password",
		ResetSent: resetSent,
		Query:     "",
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
		LoadFlash(w, r, user)

		token := r.URL.Query().Get("token")
		_, isValid := ValidatePasswordResetToken(token)

		if !isValid {
			SetFlash(w, "error", "Lien de réinitialisation invalide ou expiré.")
			http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
			return
		}

		data := struct {
			*UserInfos
			Token string
			Query string
		}{
			UserInfos: user,
			Token:     token,
			Query:     "",
		}

		tmpl, err := template.ParseFiles("pages/reset-password.html", "pages/footer.html")
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
			SetFlash(w, "error", "Lien de réinitialisation invalide ou expiré.")
			http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
			return
		}

		errorMsg := ResetPassword(email, newPassword, confirmPassword)
		if errorMsg != "" {
			SetFlash(w, "error", errorMsg)
			http.Redirect(w, r, "/reset-password?token="+url.QueryEscape(token), http.StatusSeeOther)
			return
		}

		SetFlash(w, "success", "Votre mot de passe a été réinitialisé avec succès ! Vous pouvez maintenant vous connecter.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
}

/* ===== Gestion action user =====
===== Partie post ===== */

func GetUserID(r *http.Request) int {
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
	user := GetUserFromSession(r)

	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	LoadFlash(w, r, user)

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("pages/postCreate.html", "pages/header.html", "pages/footer.html")
		data := struct {
			*UserInfos
			Page  string
			Query string
		}{
			UserInfos: user,
			Page:      "postCreate",
			Query:     "",
		}
		tmpl.ExecuteTemplate(w, "postCreate.html", data)
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

func postUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	postID, _ := strconv.Atoi(r.FormValue("id"))
	post, err := GetPostByID(postID)
	if err != nil || post.UserID != GetUserID(r) {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	newTitre := r.FormValue("titre")
	newContenu := r.FormValue("contenu")
	newCategorie := r.FormValue("categorie")

	UpdatePost(postID, newTitre, newContenu, newCategorie)
	SetFlash(w, "success", "Post mis à jour avec succès.")
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func postDelete(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if postID == 0 {
		postID, _ = strconv.Atoi(r.FormValue("id"))
	}

	post, err := GetPostByID(postID)
	if err != nil || post.UserID != GetUserID(r) {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	DeletePost(postID)
	SetFlash(w, "success", "Post supprimé.")
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func seeOnePost(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	LoadFlash(w, r, user)

	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	post, err := GetPostByID(postID)
	if err != nil {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
		return
	}

	rawComments, _ := GetComByPostID(postID)

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

	userID := GetUserID(r)
	liked := false
	if userID != 0 {
		liked = HasLiked(userID, postID)
	}

	tmpl, _ := template.ParseFiles("pages/post.html", "pages/header.html", "pages/footer.html")

	data := struct {
		Post      Post
		Comments  []ComNameAuthor
		LikeCount int
		ComCount  int
		Author    string
		Liked     bool
		*UserInfos
		Page  string
		Query string
	}{
		Post:      post,
		Comments:  comments,
		LikeCount: likeCount,
		ComCount:  comCount,
		Author:    authorPseudo,
		Liked:     liked,
		UserInfos: user,
		Page:      "post",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "post.html", data)
}

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

func seeAllPosts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	LoadFlash(w, r, user)

	posts, err := GetAllPosts()
	if err != nil {
		fmt.Println("Erreur SQL posts:", err)
		posts = []Post{}
	}

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

	tmpl, err := template.ParseFiles("pages/category.html", "pages/header.html", "pages/footer.html")
	if err != nil {
		log.Printf("Erreur chargement template category: %v", err)
		return
	}

	data := struct {
		Title string
		Posts []PostAllInfos
		*UserInfos
		Page  string
		Query string
	}{
		Title:     "Tous les posts",
		Posts:     postsInfos,
		UserInfos: user,
		Page:      "posts",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "category.html", data)
}

/*func seeAllPosts(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "<h1>Test : Le handler fonctionne !</h1>")
}*/

func myPosts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r) //recup user

	posts, _ := GetUserPosts(GetUserID(r))

	tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")

	data := struct {
		Posts []Post
		*UserInfos
		Page  string
		Query string
	}{
		Posts:     posts,
		UserInfos: user,
		Page:      "myPosts",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "account.html", data)
}

/* ===== Partie com reper ===== */

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

	http.Redirect(w, r, "/posts", http.StatusSeeOther) //secu si qlq accede en get car get pas id pas comme post donc au cas ou
}

func comUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	comID, _ := strconv.Atoi(r.FormValue("id"))
	com, err := GetComByID(comID)
	if err != nil || com.UserID != GetUserID(r) {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	newContenu := r.FormValue("contenu")
	UpdateCom(comID, newContenu)
	SetFlash(w, "success", "Commentaire mis à jour.")
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func comDelete(w http.ResponseWriter, r *http.Request) {
	comID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	com, err := GetComByID(comID)
	if err != nil || com.UserID != GetUserID(r) {
		http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
		return
	}

	DeleteCom(comID)
	SetFlash(w, "success", "Commentaire supprimé.")
	http.Redirect(w, r, "/editaccount", http.StatusSeeOther)
}

func seeMyComs(w http.ResponseWriter, r *http.Request) { //page profil user
	user := GetUserFromSession(r)

	comments, _ := GetUserCom(GetUserID(r))

	tmpl, _ := template.ParseFiles("pages/account.html", "pages/header.html", "pages/footer.html")

	data := struct {
		Comments []Com
		*UserInfos
		Page  string
		Query string
	}{
		Comments:  comments,
		UserInfos: user,
		Page:      "myComs",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "account.html", data)
}

/* ===== Partie likes ===== */

func likePost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))

	if HasLiked(userID, postID) {
		UnlikePost(userID, postID)
	} else {
		LikePost(userID, postID)
	}

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func unLikePost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))

	UnlikePost(userID, postID)

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

/* ===== Partie recherche réseau ===== */

func seeAllUsers(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp
	LoadFlash(w, r, user)
	users, _ := GetAllUsers()

	tmpl, _ := template.ParseFiles("pages/seeAllUsers.html", "pages/header.html", "pages/footer.html")

	data := struct {
		Users []User
		*UserInfos
		Page  string
		Query string
	}{
		Users:     users,
		UserInfos: user,
		Page:      "seeAllUsers",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "seeAllUsers.html", data)
}

func seeUser(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	pp, _ := getUserPP(user.Email)
	user.LoadedPP = pp

	LoadFlash(w, r, user)

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
		Page  string
		Query string
	}{
		Profile:   profile,
		Posts:     posts,
		Comments:  comments,
		UserInfos: user,
		Page:      "seeUser",
		Query:     "",
	}

	tmpl.ExecuteTemplate(w, "seeUser.html", data)
}
