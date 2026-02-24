package auth

import "golang.org/x/crypto/bcrypt"

func CheckPasswordHash(password string, passwordHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

