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
	Type string
}

type RoleWithAccess struct {
	ID            int
	Name          string
	Type          string 
	AccessedApps  sql.NullString `db:"accessed_apps"` // Menggunakan NullString untuk menampung hasil GROUP_CONCAT
}

type UserRoleDetail struct {
	Name  string         `db:"role_name"`
	Type  string         `db:"role_type"`
	Scope sql.NullString `db:"scope"`
}

// GetAllRolesForAccess mengambil semua peran untuk manajemen hak akses.
func GetAllRolesForAccess(db *sqlx.DB) ([]Role, error) {
	var roles []Role
	query := `SELECT id, name, type FROM roles ORDER BY type, name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Type); err != nil {
			continue
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// GetAllRolesWithAccess mengambil semua peran beserta aplikasi yang bisa diakses.
func GetAllRolesWithAccess(db *sqlx.DB) ([]RoleWithAccess, error) {
	var roles []RoleWithAccess
	query := `
        SELECT 
            r.id, 
            r.name, 
            r.type,
            GROUP_CONCAT(a.name SEPARATOR ', ') as accessed_apps
        FROM roles r
        LEFT JOIN application_access aa ON r.id = aa.role_id
        LEFT JOIN applications a ON aa.application_id = a.id
        GROUP BY r.id, r.name, r.type
        ORDER BY r.type, r.name`

	err := db.Select(&roles, query)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func GetUserRolesAndAttributes(db *sqlx.DB, userID int) ([]UserRoleDetail, error) {
	var roles []UserRoleDetail
	query := `SELECT r.name as role_name, r.type as role_type, ur.scope
	          FROM user_roles ur
	          JOIN roles r ON ur.role_id = r.id
	          WHERE ur.user_id = ?`
	err := db.Select(&roles, query, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	// Jika ErrNoRows, kembalikan slice kosong, bukan error
	if err == sql.ErrNoRows {
		return []UserRoleDetail{}, nil
	}
	return roles, nil
}