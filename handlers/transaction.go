package handlers

import (
	"Bank-Management-System/config"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

// Fetch user's UUID from the database using their username/email
func getUserID(username string) (string, error) {
	var userID string
	err := config.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil // No user found
	}
	return userID, err
}

// Get user's balance using their UUID
func getBalance(userID string) (int, error) {
	var balance int
	err := config.DB.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN type='deposit' THEN amount ELSE -amount END), 0)
		FROM transactions WHERE user_id=?`, userID).Scan(&balance)
	return balance, err
}

// Deposit function
func Deposit(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid deposit amount")
		return
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'deposit', ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to deposit")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Withdraw function
func Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid withdrawal amount")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	if balance < amount {
		ErrorPage(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'withdraw', ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to withdraw")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Balance function
func Balance(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}