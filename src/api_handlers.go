package forum

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func API_LikeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if !user.IsConnected {
		respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Connexion requise"})
		return
	}
	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	var dbUserID int
	db.QueryRow("SELECT id FROM Users WHERE email = ?", user.Email).Scan(&dbUserID)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM Like WHERE user_id = ? AND post_id = ?", dbUserID, postID).Scan(&count)

	if count > 0 {
		UnlikePost(dbUserID, postID)
	} else {
		LikePost(dbUserID, postID)
	}
	newCount, _ := CountLikes(postID)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"likes": newCount, "liked": count == 0})
}

func API_CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if !user.IsConnected {
		respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Non connecté"})
		return
	}
	var p Post
	json.NewDecoder(r.Body).Decode(&p)
	date := time.Now().Format("2006-01-02")
	var userID int
	db.QueryRow("SELECT id FROM Users WHERE email = ?", user.Email).Scan(&userID)

	CreatePost(p.Titre, p.Contenu, p.Categorie, date, userID)
	respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
}

func API_GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, _ := GetAllPosts()
	respondWithJSON(w, http.StatusOK, posts)
}
