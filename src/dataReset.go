package forum

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GenerateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreatePasswordResetToken(email string) (string, error) {
	token, err := GenerateResetToken()
	if err != nil {
		return "", err
	}

	var userID int
	err = db.QueryRow("SELECT id FROM Users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("email non trouvé")
	}

	expiration := time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	createdAt := time.Now().Format("2006-01-02 15:04:05")

	_, err = db.Exec("DELETE FROM PasswordReset WHERE email = ?", email)
	if err != nil {
		return "", err
	}

	_, err = db.Exec(
		"INSERT INTO PasswordReset (email, token, expiration, created_at) VALUES (?, ?, ?, ?)",
		email, token, expiration, createdAt,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func ValidatePasswordResetToken(token string) (string, bool) {
	var email string
	var expiration string

	err := db.QueryRow(
		"SELECT email, expiration FROM PasswordReset WHERE token = ?",
		token,
	).Scan(&email, &expiration)

	if err != nil {
		return "", false
	}

	expirationTime, err := time.Parse("2006-01-02 15:04:05", expiration)
	if err != nil || time.Now().After(expirationTime) {
		db.Exec("DELETE FROM PasswordReset WHERE token = ?", token)
		return "", false
	}

	return email, true
}

func SendPasswordResetEmail(email string, token string) error {
	resetLink := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)
	fmt.Printf("\n[PASSWORD RESET EMAIL]\nTo: %s\n\n", email)
	fmt.Printf("Cliquez sur le lien ci-dessous pour réinitialiser votre mot de passe:\n%s\n\n", resetLink)

	/*
		//Pour la prod
		from := "noreply@forum.com"
		password := "your_email_password"
		to := []string{email}
		smtpHost := "smtp.gmail.com"
		smtpPort := "587"
		message := fmt.Sprintf("Subject: Réinitialisation de mot de passe\n\n"+
		    "Cliquez sur le lien ci-dessous pour réinitialiser votre mot de passe:\n%s\n"+
		    "Ce lien expire dans 1 heure.", resetLink)
		auth := smtp.PlainAuth("", from, password, smtpHost)
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
		return err
	*/

	return nil
}

func ResetPassword(email string, newPassword string, confirmPassword string) string {
	if newPassword != confirmPassword {
		return "Les mots de passe ne correspondent pas."
	}

	if len(newPassword) < 12 {
		return "La taille du mot de passe doit être d'au moins 12 caractères."
	}

	var allowedCharacters = regexp.MustCompile(`^[\x21-\x7E]+$`)
	if !allowedCharacters.MatchString(newPassword) {
		return "Le mot de passe contient des caractères non autorisés."
	}

	var CPC_hasUpper = regexp.MustCompile(`[A-Z]`)
	var CPC_hasLower = regexp.MustCompile(`[a-z]`)
	var CPC_hasDigit = regexp.MustCompile(`[0-9]`)
	var hasSpecial = regexp.MustCompile(`[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`)

	if !CPC_hasUpper.MatchString(newPassword) ||
		!CPC_hasLower.MatchString(newPassword) ||
		!CPC_hasDigit.MatchString(newPassword) ||
		!hasSpecial.MatchString(newPassword) {
		return "La composition du mot de passe ne respecte pas les critères attendus."
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "Erreur lors du traitement du mot de passe."
	}

	_, err = db.Exec(
		"UPDATE Users SET password_hash = ? WHERE email = ?",
		string(hash), email,
	)
	if err != nil {
		return "Erreur lors de la mise à jour du mot de passe."
	}

	_, err = db.Exec("DELETE FROM PasswordReset WHERE email = ?", email)
	if err != nil {
		fmt.Println("Erreur lors de la suppression du token:", err)
	}

	return ""
}

func DataForgotPasswordSend(w http.ResponseWriter, r *http.Request, email string) string {
	email = strings.TrimSpace(email)

	if email == "" {
		return "Veuillez entrer une adresse email."
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "Adresse email invalide. Veuillez réessayer."
	}

	var userID int
	err := db.QueryRow("SELECT id FROM Users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		return "Si cet email existe, un lien de réinitialisation a été envoyé."
	}

	token, err := CreatePasswordResetToken(email)
	if err != nil {
		fmt.Println("Erreur création token:", err)
		return "Erreur lors de la création du token. Veuillez réessayer."
	}

	err = SendPasswordResetEmail(email, token)
	if err != nil {
		fmt.Println("Erreur envoi email:", err)
	}

	return "Si cet email existe, un lien de réinitialisation a été envoyé."
}
