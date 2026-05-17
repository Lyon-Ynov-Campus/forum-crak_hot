package forum

import (
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("DB non initialisée")
	}
	return db, nil
}

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "Forum.db")
	if err != nil {
		panic(err)
	}
}

func CreateDB() {
	InitDB()

	schema := `
	CREATE TABLE IF NOT EXISTS Users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		photo_profil TEXT 
	);

	CREATE TABLE IF NOT EXISTS Post (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		titre TEXT NOT NULL,
		contenu TEXT NOT NULL,
		categorie TEXT,
		date_publication TEXT,
		user_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS Commentaire (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		contenu TEXT NOT NULL,
		date_com TEXT,
		user_id INTEGER,
		post_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE,
		FOREIGN KEY(post_id) REFERENCES Post(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS Like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		post_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE,
		FOREIGN KEY(post_id) REFERENCES Post(id) ON DELETE CASCADE,
		UNIQUE(user_id, post_id)
	);

	CREATE TABLE IF NOT EXISTS Session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		token TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS PasswordReset (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expiration TEXT NOT NULL,
		created_at TEXT NOT NULL
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		fmt.Println("Erreur lors de la création des tables:", err)
		panic(err)
	}
}

type User struct {
	ID          int    `json:"id"`
	Pseudo      string `json:"pseudo"`
	Email       string `json:"email"`
	MotDePasse  string `json:"mot_de_passe"`
	PhotoProfil string `json:"photo_profil"`
}

type Post struct {
	ID              int    `json:"id"`
	Titre           string `json:"titre"`
	Contenu         string `json:"contenu"`
	Categorie       string `json:"categorie"`
	DatePublication string `json:"date_publication"`
	UserID          int    `json:"user_id"`
	CountLikes      int    `json:"count_likes"`
	CountCom        int    `json:"count_com"`
}

type Com struct {
	ID      int    `json:"id"`
	Contenu string `json:"contenu"`
	DateCom string `json:"date_com"`
	UserID  int    `json:"user_id"`
	PostID  int    `json:"post_id"`
}

func CreateUser(pseudo, email, motDePasse string) error {
	insertQuery := `
        INSERT INTO User(pseudo, email, mot_de_passe)
        VALUES(?, ?, ?)
    `
	_, err := db.Exec(insertQuery, pseudo, email, motDePasse)
	return err
}

func GetUserByID(id int) (User, error) {
	var u User
	row := db.QueryRow(`
        SELECT id, pseudo, email, mot_de_passe, photo_profil
        FROM User
        WHERE id = ?
    `, id)

	err := row.Scan(&u.ID, &u.Pseudo, &u.Email, &u.MotDePasse, &u.PhotoProfil)
	return u, err
}

/* ===== Partie update de la page profil user ===== */

func UpdateUEmail(id int, newEmail string) error {
	_, err := db.Exec("UPDATE Users SET email = ? WHERE id = ?", newEmail, id)
	return err
}

func UpdateUPseudo(id int, newPseudo string) error {
	_, err := db.Exec("UPDATE Users SET username = ? WHERE id = ?", newPseudo, id)
	return err
}

func UpdateUPassword(id int, newHash string) error {
	_, err := db.Exec("UPDATE Users SET password_hash = ? WHERE id = ?", newHash, id)
	return err
}

func UpdateUPhoto(id int, photo string) error {
	updateQuery := `
        UPDATE User
        SET photo_profil = ?
        WHERE id = ?
    `
	_, err := db.Exec(updateQuery, photo, id)
	return err
}

func GetUserPosts(userID int) ([]Post, error) {
	rows, err := db.Query("SELECT id, titre, contenu, date_publication, user_id FROM Post WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Titre, &p.Contenu, &p.DatePublication, &p.UserID)
		posts = append(posts, p)
	}
	return posts, nil
}

func GetUserCom(userID int) ([]Com, error) {
	rows, err := db.Query(`
        SELECT id, contenu, date_com, user_id, post_id
        FROM Commentaire
        WHERE user_id = ?
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Com
	for rows.Next() {
		var c Com
		rows.Scan(&c.ID, &c.Contenu, &c.DateCom, &c.UserID, &c.PostID)
		comments = append(comments, c)
	}
	return comments, nil
}

func DeleteUser(id int) error { //supprimer compte avec tout data
	deleteQuery := `
        DELETE FROM User
        WHERE id = ?
    `
	_, err := db.Exec(deleteQuery, id)
	return err
}

/* ===== Partie posts =====*/

func CreatePost(titre, contenu, categorie, date string, userID int) error {
	_, err := db.Exec("INSERT INTO Post (titre, contenu, categorie, date_publication, user_id) VALUES (?, ?, ?, ?, ?)",
		titre, contenu, categorie, date, userID)
	return err
}

func GetPostByID(id int) (Post, error) {
	var p Post
	err := db.QueryRow("SELECT id, titre, contenu, categorie, date_publication, user_id FROM Post WHERE id = ?", id).Scan(&p.ID, &p.Titre, &p.Contenu, &p.Categorie, &p.DatePublication, &p.UserID)
	return p, err
}

func UpdatePost(id int, newTitre, newContenu, newCategorie string) error {
	_, err := db.Exec("UPDATE Post SET titre = ?, contenu = ?, categorie = ? WHERE id = ?", newTitre, newContenu, newCategorie, id)
	return err
}

func DeletePost(id int) error {
	db.Exec("DELETE FROM Commentaire WHERE post_id = ?", id)
	_, err := db.Exec("DELETE FROM Post WHERE id = ?", id)
	return err
}

func GetAllPosts() ([]Post, error) {
	rows, err := db.Query("SELECT id, titre, contenu, categorie, date_publication, user_id FROM Post ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Titre, &p.Contenu, &p.Categorie, &p.DatePublication, &p.UserID)
		posts = append(posts, p)
	}
	return posts, nil
}

func GetPseudoByUserID(id int) (string, error) {
	var pseudo string
	err := db.QueryRow("SELECT username FROM Users WHERE id = ?", id).Scan(&pseudo)
	return pseudo, err
}

/* ===== Patie commentaires ===== */

func CreateCom(contenu, dateCom string, userID, postID int) error {
	_, err := db.Exec("INSERT INTO Commentaire (contenu, date_com, user_id, post_id) VALUES (?, ?, ?, ?)",
		contenu, dateCom, userID, postID)
	return err
}

func UpdateCom(id int, newContenu string) error {
	updateQuery := `
        UPDATE Commentaire
        SET contenu = ?
        WHERE id = ?
    `
	_, err := db.Exec(updateQuery, newContenu, id)
	return err
}

func DeleteCom(id int) error {
	deleteQuery := `
        DELETE FROM Commentaire
        WHERE id = ?
    `
	_, err := db.Exec(deleteQuery, id)
	return err
}

func GetComByPostID(postID int) ([]Com, error) {
	rows, err := db.Query("SELECT id, contenu, date_com, user_id, post_id FROM Commentaire WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Com
	for rows.Next() {
		var c Com
		rows.Scan(&c.ID, &c.Contenu, &c.DateCom, &c.UserID, &c.PostID)
		comments = append(comments, c)
	}
	return comments, nil
}

func CountCom(postID int) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Commentaire WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func GetComByID(id int) (Com, error) {
	var c Com

	row := db.QueryRow(`
        SELECT id, contenu, date_com, user_id, post_id
        FROM Commentaire
        WHERE id = ?
    `, id)

	err := row.Scan(&c.ID, &c.Contenu, &c.DateCom, &c.UserID, &c.PostID)
	return c, err
}

/* ===== Partie likes ===== */

func LikePost(userID, postID int) error {
	_, err := db.Exec("INSERT INTO Like (user_id, post_id) VALUES (?, ?)", userID, postID)
	return err
}

func UnlikePost(userID, postID int) error {
	_, err := db.Exec("DELETE FROM Like WHERE user_id = ? AND post_id = ?", userID, postID)
	return err
}

func CountLikes(postID int) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Like WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func HasLiked(userID, postID int) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Like WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&count)
	return err == nil && count > 0
}

/* ===== Partie recherche =====*/

func SearchPostsByTitle(query string) ([]Post, error) {
	rows, err := db.Query("SELECT id, titre, contenu, categorie, date_publication, user_id FROM Post WHERE titre LIKE ?", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Titre, &p.Contenu, &p.Categorie, &p.DatePublication, &p.UserID)
		posts = append(posts, p)
	}
	return posts, nil
}

/* ===== Partie réseau ===== */

func GetAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, email, username, photo_profil FROM Users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Email, &u.Pseudo, &u.PhotoProfil)
		users = append(users, u)
	}
	return users, nil
}

func SearchUsersByName(pseudo string) ([]User, error) {
	rows, err := db.Query("SELECT id, email, username, photo_profil FROM Users WHERE username LIKE ?", "%"+pseudo+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Email, &u.Pseudo, &u.PhotoProfil)
		users = append(users, u)
	}
	return users, nil
}

func ApplyTri(posts []Post, sortPar string) []Post {
	switch sortPar {
	case "date_asc":
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].DatePublication < posts[j].DatePublication
		})
	case "date_desc":
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].DatePublication > posts[j].DatePublication
		})
	case "likes_asc":
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CountLikes < posts[j].CountLikes
		})
	case "likes_desc":
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CountLikes > posts[j].CountLikes
		})
	case "com_asc":
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CountCom < posts[j].CountCom
		})
	case "com_desc":
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CountCom > posts[j].CountCom
		})
	}
	return posts
}

func getUserPP(email string) (string, error) {
	var pp string
	err := db.QueryRow("SELECT photo_profil FROM Users WHERE email = ?", email).Scan(&pp)
	return pp, err
}

func updateUserPP(email string, ppURL string) error {
	_, err := db.Exec("UPDATE Users SET photo_profil = ? WHERE email = ?", ppURL, email)
	return err
}
