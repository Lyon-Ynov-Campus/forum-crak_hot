package forum

import (
	//"fmt"

	"database/sql"

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
