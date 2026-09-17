package migrations

import "embed"

// FS holds SQL migration files for golang-migrate.
//
//go:embed *.sql
var FS embed.FS
