package forum

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// respondWithJSON est une fonction utilitaire pour envoyer des réponses au format JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// API_LikeHandler gère l'ajout et la suppression de likes via l'API
func API_LikeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if !user.IsConnected {
		respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Connexion requise"})
		return
	}

	postID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	userID := GetUserID(r)

	// Vérifier si l'utilisateur a déjà liké ce post
	var count int
	db.QueryRow("SELECT COUNT(*) FROM Like WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&count)

	if count > 0 {
		// Si déjà liké, on retire le like
		UnlikePost(userID, postID)
	} else {
		// Sinon, on ajoute le like
		LikePost(userID, postID)
	}

	// Récupérer le nouveau compteur total de likes
	newCount, _ := CountLikes(postID)

	// Répondre avec le nouveau total et l'état actuel (si l'utilisateur vient de liker ou non)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"likes": newCount,
		"liked": count == 0,
	})
}

// API_CreatePostHandler permet de créer un nouveau post via une requête JSON (fetch)
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

	// Appel de la fonction SQL de db.go
	err = CreatePost(p.Titre, p.Contenu, p.Categorie, date, userID)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erreur lors de la création du post"})
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
}

// API_GetPostsHandler renvoie la liste complète des posts au format JSON
func API_GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := GetAllPosts()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erreur lors de la récupération des posts"})
		return
	}

	// Renvoie la liste des posts (très utile pour vérifier que l'API fonctionne)
	respondWithJSON(w, http.StatusOK, posts)
}
