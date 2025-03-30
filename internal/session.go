package internal

import (
	"os"
	"strings"
)

var CurrentUserID string

func IsLoggedIn() bool {
	return GetSessionUser() != ""
}

func GetSessionUser() string {
	data, err := os.ReadFile(".session")
	if err != nil {
		return ""
	}
	CurrentUserID = strings.TrimSpace(string(data))
	return CurrentUserID
}

func SaveSession(empID string) error {
	CurrentUserID = empID
	return os.WriteFile(".session", []byte(empID), 0644)
}

func ClearSession() error {
	CurrentUserID = ""
	return os.Remove(".session")
}
