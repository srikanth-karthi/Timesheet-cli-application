package setup

import (
	"os"

	"github.com/srikanth-karthi/timesheet/internal/auth"
)

func GetCredentialProvider() auth.CredentialProvider {
	if _, err := os.Stat("credentials.json"); err == nil {
		return auth.FileProvider{Path: "credentials.json"}
	}
	return auth.EmbeddedProvider{}
}