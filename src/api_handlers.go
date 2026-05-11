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
	userID := GetUserID(r)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM Like WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&count)

	if count > 0 {
		UnlikePost(userID, postID)
	} else {
		LikePost(userID, postID)
	}

	newCount, _ := CountLikes(postID)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"likes": newCount,
		"liked": count == 0,
	})
}

func API_CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if !user.IsConnected {
		respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Non connecté"})
		return
	}

	var p Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Données invalides"})
		return
	}

	userID := GetUserID(r)
	date := time.Now().Format("2006-01-02")

	err = CreatePost(p.Titre, p.Contenu, p.Categorie, date, userID)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erreur lors de la création du post"})
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
}

func API_GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := GetAllPosts()
    if err != nil {
        respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erreur SQL"})
        return
    }
	if posts == nil {
        posts = []Post{} 
    }
    respondWithJSON(w, http.StatusOK, posts)
}
