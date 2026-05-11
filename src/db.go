package forum

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

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
		FOREIGN KEY(user_id) REFERENCES Users(id)
	);

	CREATE TABLE IF NOT EXISTS Commentaire (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		contenu TEXT NOT NULL,
		date_com TEXT,
		user_id INTEGER,
		post_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES Users(id),
		FOREIGN KEY(post_id) REFERENCES Post(id)
	);

	CREATE TABLE IF NOT EXISTS Like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		post_id INTEGER,
		UNIQUE(user_id, post_id)
	);

	CREATE TABLE IF NOT EXISTS Session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		token TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES Users(id)
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
	MotDePasse string `json:"mot_de_passe"`
	PhotoProfil string `json:"photo_profil"`
}

type Post struct {
	ID              int    `json:"id"`
	Titre           string    `json:"titre"`
	Contenu         string    `json:"contenu"`
	Categorie       string    `json:"categorie"`
	DatePublication string    `json:"date_publication"`
	UserID          int    `json:"user_id"`
}

type Com struct {
	ID      int    `json:"id"`
	Contenu string `json:"contenu"`
	DateCom string `json:"date_com"`
	UserID  int    `json:"user_id"`
	PostID  int    `json:"post_id"`
}

func GetUserByID(id int) (User, error) {
	var u User
	err := db.QueryRow("SELECT id, username, email, photo_profil FROM Users WHERE id = ?", id).Scan(&u.ID, &u.Pseudo, &u.Email, &u.PhotoProfil)
	return u, err
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

func CreatePost(titre, contenu, categorie, date string, userID int) error {
	_, err := db.Exec("INSERT INTO Post (titre, contenu, categorie, date_publication, user_id) VALUES (?, ?, ?, ?, ?)",
		titre, contenu, categorie, date, userID)
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

func CreateCom(contenu, dateCom string, userID, postID int) error {
	_, err := db.Exec("INSERT INTO Commentaire (contenu, date_com, user_id, post_id) VALUES (?, ?, ?, ?)",
		contenu, dateCom, userID, postID)
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

func AddLike(userID, postID int) error {
	_, err := db.Exec("INSERT INTO Like (user_id, post_id) VALUES (?, ?)", userID, postID)
	return err
}

func RemoveLike(userID, postID int) error {
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

func GetPostByID(id int) (Post, error) {
	var p Post
	err := db.QueryRow("SELECT id, titre, contenu, categorie, date_publication, user_id FROM Post WHERE id = ?", id).Scan(&p.ID, &p.Titre, &p.Contenu, &p.Categorie, &p.DatePublication, &p.UserID)
	return p, err
}

func GetPseudoByUserID(id int) (string, error) {
	var pseudo string
	err := db.QueryRow("SELECT username FROM Users WHERE id = ?", id).Scan(&pseudo)
	return pseudo, err
}

func LikePost(userID, postID int) error {
	_, err := db.Exec("INSERT INTO Like (user_id, post_id) VALUES (?, ?)", userID, postID)
	return err
}

func UnlikePost(userID, postID int) error {
	_, err := db.Exec("DELETE FROM Like WHERE user_id = ? AND post_id = ?", userID, postID)
	return err
}