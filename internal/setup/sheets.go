package setup

import (
	"context"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/srikanth-karthi/timesheet/internal/auth"
)

func GetSheetsService(provider auth.CredentialProvider) *sheets.Service {
	creds, err := provider.GetJSON()
	if err != nil {
		log.Fatalf("  Failed to load credentials: %v", err)
	}

	srv, err := sheets.NewService(context.Background(), option.WithCredentialsJSON(creds))
	if err != nil {
		log.Fatalf("  Unable to create Sheets service: %v", err)
	}

	return srv
}
