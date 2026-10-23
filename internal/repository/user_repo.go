package repository

import (
	"portal/internal/domain"
)

var (
	users = make(map[string]domain.PasswordHash)
)

func SaveUser(username string, password domain.PasswordHash) {
	users[username] = password
}

func GetUser(username string) (domain.PasswordHash, bool) {
	password, exists := users[username]
	
	return password, exists
}