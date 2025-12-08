// file: models/role.go

package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Role adalah struct untuk data peran.
type Role struct {
	ID   int `db:"id"`
	Name string `db:"role_name"`
	Description string `db:"description"`
}

type RoleWithAccess struct {
	ID           int
	Name         string
	AccessedApps sql.NullString `db:"accessed_apps"`
}

type DCRole struct {
	ID          int     `json:"id_role"`
	Name        string  `json:"nama_role"`
	Description string  `json:"description"`
	DeletedAt   *string `json:"deleted_at"`
}

type DCRoleResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   []DCRole `json:"data"`
}

func GetAllRoles(db *sqlx.DB) ([]Role, error) {
	var data []Role
	err := db.Select(&data, "SELECT id, role_name FROM roles ORDER BY id ASC")
	return data, err
}

func UpsertRole(db *sqlx.DB, data DCRole) error {
	query := `
		INSERT INTO roles 
			(id, role_name, description, created_at, updated_at, deleted_at)
		VALUES 
			(:id, :name, :desc, NOW(), NOW(), :deleted_at)
		ON DUPLICATE KEY UPDATE
			role_name = VALUES(role_name),
			description = VALUES(description),
			updated_at = NOW(),
			deleted_at = VALUES(deleted_at);
	`
	params := map[string]interface{}{
		"id":         data.ID,
		"name":       data.Name,
		"desc":       data.Description,
		"deleted_at": data.DeletedAt,
	}
	_, err := db.NamedExec(query, params)
	return err
}
