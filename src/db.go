package forum

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" //ps oublier import github.com voir repo soutien
)

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
	CREATE TABLE IF NOT EXISTS User(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	pseudo TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	mot_de_passe TEXT NOT NULL,
	photo_profil TEXT 
	);
	` //type TEXT pr photo car soit nom du file soit url de la P

	_, err := db.Exec(CreateTableUser)
	if err != nil {
		fmt.Println("erreur table user", err)
		panic(err)
	}

	CreateTablePost := `
	CREATE TABLE IF NOT EXISTS Post(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	titre TEXT NOT NULL,
	contenu TEXT NOT NULL,
	categorie TEXT NOT NULL,
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
	FOREIGN KEY (post_id) REFERENCES Post(id) ON DELETE CASCADE,
	UNIQUE(user_id, post_id)
	);
	` //rev UNIQUE de w3scool pr pas que user like 2 fois

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
	defer db.Close()

}

type User struct {
	ID          int
	Pseudo      string
	Email       string
	MotDePasse  string
	PhotoProfil string //a voir car optionnel
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

// rappel dysca : ? val a fournir + tard
// PARTIE USER

func CreateUser(pseudo, email, motDePasse string) error { //creer profil
	insertQuery := `
        INSERT INTO User(pseudo, email, mot_de_passe)
        VALUES(?, ?, ?)
    `
	_, err := db.Exec(insertQuery, pseudo, email, motDePasse)
	return err
}

func GetUserByID(id int) (User, error) { //recup user par id car si on veut afficher info d'un user on doit recup les infos via id
	var u User
	row := db.QueryRow(`
        SELECT id, pseudo, email, mot_de_passe, photo_profil
        FROM User
        WHERE id = ?
    `, id)

	err := row.Scan(&u.ID, &u.Pseudo, &u.Email, &u.MotDePasse, &u.PhotoProfil)
	return u, err
}

//partie update de la page profil user

func UpdateUEmail(id int, newEmail string) error {
	updateQuery := `
        UPDATE User
        SET email = ?
        WHERE id = ?
    `
	_, err := db.Exec(updateQuery, newEmail, id)
	return err
}

func UpdateUPseudo(id int, newPseudo string) error {
	updateQuery := `
        UPDATE User
        SET pseudo = ?
        WHERE id = ?
    `
	_, err := db.Exec(updateQuery, newPseudo, id)
	return err
}

func UpdateUPassword(id int, newPassword string) error {
	updateQuery := `
        UPDATE User
        SET mot_de_passe = ?
        WHERE id = ?
    `
	_, err := db.Exec(updateQuery, newPassword, id)
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
	rows, err := db.Query(`
        SELECT id, titre, contenu, date_publication, user_id
        FROM Post
        WHERE user_id = ?
    `, userID)
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

//PARITE post

func CreatePost(titre, contenu, categorie, datePublication string, userID int) error {
	insertQuery := `
        INSERT INTO Post(titre, contenu, categorie, date_publication, user_id)
        VALUES(?, ?, ?, ?, ?)
    `
	_, err := db.Exec(insertQuery, titre, contenu, categorie, datePublication, userID)
	return err
}

func GetPostByID(id int) (Post, error) { //recup 1 seul post grace a son id
	var p Post

	row := db.QueryRow(`
        SELECT id, titre, contenu, categorie, date_publication, user_id
        FROM Post
        WHERE id = ?
    `, id)

	err := row.Scan(&p.ID, &p.Titre, &p.Contenu, &p.Categorie, &p.DatePublication, &p.UserID)
	return p, err
}

func UpdatePost(id int, newTitre, newContenu, newCategorie string) error {
	updateQuery := `
        UPDATE Post
        SET titre = ?, contenu = ?, categorie = ?
        WHERE id = ?
    `
	_, err := db.Exec(updateQuery, newTitre, newContenu, newCategorie, id)
	return err
}

func DeletePost(id int) error {
	deleteQuery := `
        DELETE FROM Post
        WHERE id = ?
    `
	_, err := db.Exec(deleteQuery, id)
	return err
}

func GetAllPosts() ([]Post, error) {
	rows, err := db.Query(`
        SELECT id, titre, contenu, categorie, date_publication, user_id
        FROM Post
    `)
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

func GetPseudoByUserID(id int) (string, error) { //affiche autzur d'un post
	var pseudo string
	row := db.QueryRow(`
        SELECT pseudo
        FROM User
        WHERE id = ?
    `, id)

	err := row.Scan(&pseudo)
	return pseudo, err
}

// PARTIE Commetnaire

func CreateCom(contenu, dateCom string, userID, postID int) error {
	insertQuery := `
        INSERT INTO Commentaire(contenu, date_com, user_id, post_id)
        VALUES(?, ?, ?, ?)
    `
	_, err := db.Exec(insertQuery, contenu, dateCom, userID, postID)
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

func GetComByPostID(postID int) ([]Com, error) { //recup tout les comm d'un post precis
	rows, err := db.Query(`
        SELECT id, contenu, date_com, user_id, post_id
        FROM Commentaire
        WHERE post_id = ?
    `, postID)
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

	row := db.QueryRow(`
        SELECT COUNT(*)
        FROM Commentaire
        WHERE post_id = ?
    `, postID)

	err := row.Scan(&count)
	return count, err
}

//PARTIE Like

func LikePost(userID, postID int) error {
	insertQuery := `
        INSERT INTO Like(user_id, post_id)
        VALUES(?, ?)
    `
	_, err := db.Exec(insertQuery, userID, postID)
	return err
}

func UnlikePost(userID, postID int) error {
	deleteQuery := `
        DELETE FROM Like
        WHERE user_id = ? AND post_id = ?
    `
	_, err := db.Exec(deleteQuery, userID, postID)
	return err
}

func CountLikes(postID int) (int, error) {
	var count int

	row := db.QueryRow(`
        SELECT COUNT(*)
        FROM Like
        WHERE post_id = ?
    `, postID)

	err := row.Scan(&count)
	return count, err
}

//Partie recherche

func SearchPostsByTitle(r string) ([]Post, error) {
	rows, err := db.Query(`
        SELECT id, titre, contenu, categorie, date_publication, user_id
        FROM Post
        WHERE titre LIKE ?
    `, "%"+r+"%") //le %% c tout les text qui contient x mot ligne faite par IA
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

//partie reseau

func GetAllUsers() ([]User, error) {
	rows, err := db.Query(`
        SELECT id,email, pseudo, photo_profil
        FROM User
    `)
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
	rows, err := db.Query(`
        SELECT id,email, pseudo, photo_profil
        FROM User
        WHERE pseudo LIKE ?
    `, "%"+pseudo+"%") //ligne IA
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
