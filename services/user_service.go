package services

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"netflix_central/database"
	"netflix_central/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

func CreateUser(email, password string) (*models.User, error) {
	db := database.GetDB()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var id int64
	err = db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, string(hash)).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &models.User{ID: id, Email: email, PasswordHash: string(hash)}, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	db := database.GetDB()
	row := db.QueryRow("SELECT id, email, password_hash FROM users WHERE email = $1", email)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func ValidateUser(email, password string) (*models.User, error) {
	u, err := GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}
