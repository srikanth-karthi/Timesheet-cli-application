package auth

import (
	"os"

	"github.com/srikanth-karthi/timesheet/internal/creds"
)

type CredentialProvider interface {
	GetJSON() ([]byte, error)
}

type FileProvider struct {
	Path string
}

func (fp FileProvider) GetJSON() ([]byte, error) {
	return os.ReadFile(fp.Path)
}

type EmbeddedProvider struct{}

func (ep EmbeddedProvider) GetJSON() ([]byte, error) {
	return creds.EmbeddedCreds, nil
}
