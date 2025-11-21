// file: models/role.go

package models

import (
	"database/sql"
)

// Role adalah struct untuk data peran.
type Role struct {
	ID   int
	Name string
}

type RoleWithAccess struct {
	ID            int
	Name          string
	AccessedApps  sql.NullString `db:"accessed_apps"`
}