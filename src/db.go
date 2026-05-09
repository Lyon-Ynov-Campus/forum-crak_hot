package forum

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func OpenDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("DB non initialisée")
	}
	return db, nil
}

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "Forum.db")
	if err != nil {
		panic(err)
	}
}

func CreateDB() {
	InitDB()
	db.Exec(`CREATE TABLE IF NOT EXISTS Users(id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, photo_profil TEXT);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS Post(id INTEGER PRIMARY KEY AUTOINCREMENT, titre TEXT NOT NULL, contenu TEXT NOT NULL, categorie TEXT NOT NULL, date_publication TEXT NOT NULL, user_id INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS Commentaire(id INTEGER PRIMARY KEY AUTOINCREMENT, contenu TEXT NOT NULL, date_com TEXT NOT NULL, user_id INTEGER NOT NULL, post_id INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE, FOREIGN KEY (post_id) REFERENCES Post(id) ON DELETE CASCADE);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS Like(id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, post_id INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE, FOREIGN KEY (post_id) REFERENCES Post(id) ON DELETE CASCADE, UNIQUE(user_id, post_id));`)
	db.Exec(`CREATE TABLE IF NOT EXISTS Session(id INTEGER PRIMARY KEY AUTOINCREMENT, token TEXT NOT NULL, user_id INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS PasswordReset(id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL, token TEXT NOT NULL UNIQUE, expiration TEXT NOT NULL, created_at TEXT NOT NULL);`)
}

type User struct {
	ID          int
	Pseudo      string
	Email       string
	MotDePasse  string
	PhotoProfil string
}

type Post struct {
	ID              int
	Titre           string
	Contenu         string
	Categorie       string
	DatePublication string
	UserID          int
}

type Com struct {
	ID      int
	Contenu string
	DateCom string
	UserID  int
	PostID  int
}

func CreatePost(titre, contenu, categorie, datePublication string, userID int) error {
	_, err := db.Exec("INSERT INTO Post(titre, contenu, categorie, date_publication, user_id) VALUES(?, ?, ?, ?, ?)", titre, contenu, categorie, datePublication, userID)
	return err
}

func GetPostByID(id int) (Post, error) {
	var p Post
	err := db.QueryRow("SELECT id, titre, contenu, categorie, date_publication, user_id FROM Post WHERE id = ?", id).Scan(&p.ID, &p.Titre, &p.Contenu, &p.Categorie, &p.DatePublication, &p.UserID)
	return p, err
}

func GetAllPosts() ([]Post, error) {
	rows, err := db.Query("SELECT id, titre, contenu, categorie, date_publication, user_id FROM Post")
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

func GetPseudoByUserID(id int) (string, error) {
	var pseudo string
	err := db.QueryRow("SELECT username FROM Users WHERE id = ?", id).Scan(&pseudo)
	return pseudo, err
}

func CountLikes(postID int) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Like WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func CountCom(postID int) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Commentaire WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func LikePost(userID, postID int) error {
	_, err := db.Exec("INSERT INTO Like(user_id, post_id) VALUES(?, ?)", userID, postID)
	return err
}

func UnlikePost(userID, postID int) error {
	_, err := db.Exec("DELETE FROM Like WHERE user_id = ? AND post_id = ?", userID, postID)
	return err
}

func getUserPP(email string) (string, error) {
	var pp string
	err := db.QueryRow("SELECT photo_profil FROM Users WHERE email = ?", email).Scan(&pp)
	return pp, err
}

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

func GetUserByID(id int) (User, error) {
	var u User
	err := db.QueryRow("SELECT id, username, email, photo_profil FROM Users WHERE id = ?", id).Scan(&u.ID, &u.Pseudo, &u.Email, &u.PhotoProfil)
	return u, err
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
	rows, err := db.Query("SELECT id, contenu, date_com, user_id, post_id FROM Commentaire WHERE user_id = ?", userID)
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

func CreateCom(contenu, dateCom string, userID, postID int) error {
	_, err := db.Exec("INSERT INTO Commentaire(contenu, date_com, user_id, post_id) VALUES(?, ?, ?, ?)", contenu, dateCom, userID, postID)
	return err
}

func UpdatePost(id int, titre, contenu, categorie string) error {
	_, err := db.Exec("UPDATE Post SET titre = ?, contenu = ?, categorie = ? WHERE id = ?", titre, contenu, categorie, id)
	return err
}

func DeletePost(id int) error {
	_, err := db.Exec("DELETE FROM Post WHERE id = ?", id)
	return err
}
