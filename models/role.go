// file: models/role.go

package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Role adalah struct untuk data peran.
type Role struct {
	ID   int
	Name string
}

type RoleWithAccess struct {
	ID           int
	Name         string
	AccessedApps sql.NullString `db:"accessed_apps"`
}

func GetAllRoles(db *sqlx.DB) ([]Role, error) {

	var role []Role
	query := `SELECT id, role_name AS name FROM roles`

	err := db.Select(&role, query)
	return role,err
}
