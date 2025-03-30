package creds

import _ "embed"

//go:embed credentials.json
var EmbeddedCreds []byte
