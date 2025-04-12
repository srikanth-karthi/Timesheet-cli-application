package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Meta struct {
	Active       string `json:"active"`
	SessionStart string `json:"session_start"`
}

var metaPath = filepath.Join(os.Getenv("HOME"), ".timesheet", "meta.json")

func LoadMeta() (*Meta, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return &Meta{Active: "general"}, nil
	}
	var meta Meta
	err = json.Unmarshal(data, &meta)
	return &meta, err
}

func SaveMeta(meta *Meta) error {
	os.MkdirAll(filepath.Dir(metaPath), 0755)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}
