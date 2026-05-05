package forum

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" //ps oublier import github.com voir repo soutien
)

func OpenDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("DB non initialisée")
	}
	return db, nil
}

var db *sql.DB

func InitDB() { //corps debut repo soutien rev
	var err error
	db, err = sql.Open("sqlite3", "Forum.db")
	if err != nil {
		panic(err)
	}

}

func CreateDB() { //rev slide 39 soutien pour creer table
	InitDB()
	CreateTableUser := `
	CREATE TABLE IF NOT EXISTS Users(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	photo_profil TEXT 
	);
	` //type TEXT pr photo car soit nom du file soit url de la P

	_, err := db.Exec(CreateTableUser)
	if err != nil {
		fmt.Println("erreur table users", err)
		panic(err)
	}

	CreateTablePost := `
	CREATE TABLE IF NOT EXISTS Post(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	titre TEXT NOT NULL,
	contenu TEXT NOT NULL,
	date_publication TEXT NOT NULL,
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES User(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(CreateTablePost)
	if err != nil {
		fmt.Println("erreur table post", err)
		panic(err)
	}

	CreateTableCommentaire := `
	CREATE TABLE IF NOT EXISTS Commentaire(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	contenu TEXT NOT NULL,
	date_com TEXT NOT NULL,
	user_id INTEGER NOT NULL,
	post_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES User(id) ON DELETE CASCADE,
	FOREIGN KEY (post_id) REFERENCES Post(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(CreateTableCommentaire)
	if err != nil {
		fmt.Println("erreur table commentaire", err)
		panic(err)
	}

	CreateTableLike := `
	CREATE TABLE IF NOT EXISTS Like(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	post_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES User(id) ON DELETE CASCADE,
	FOREIGN KEY (post_id) REFERENCES Post(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(CreateTableLike) //att rappel var deja creer donc pas remmettre := mais =
	if err != nil {
		fmt.Println("erreur table Like", err)
		panic(err)
	}

	CreateTableSession := `
	CREATE TABLE IF NOT EXISTS Session(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	token TEXT NOT NULL,
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES User(id) ON DELETE CASCADE
	);
	` //rev doc datacamp pour ON DELETE CASCADE pour consigne effacer data si compte suppr

	_, err = db.Exec(CreateTableSession)
	if err != nil {
		fmt.Println("erreur table session", err)
		panic(err)
	}

}
