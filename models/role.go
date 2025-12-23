// file: models/role.go

package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

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

func GetAllRoles(db *sqlx.DB) ([]Role, error) {
	var data []Role
	err := db.Select(&data, "SELECT id, role_name, description FROM roles WHERE deleted_at IS NULL ORDER BY id ASC")
	return data, err
}

func FindRoleByID(db *sqlx.DB, id int) (*Role, error) {
	var r Role
	// Ambil role yang belum dihapus (jika pakai soft delete)
	// Atau ambil raw jika hard delete. Kita asumsi soft delete sesuai pola sebelumnya.
	query := "SELECT id, role_name, description FROM roles WHERE id = ? AND deleted_at IS NULL"
	err := db.Get(&r, query, id)
	return &r, err
}

func CreateRole(db *sqlx.DB, name, description string) error {
	query := `INSERT INTO roles (role_name, description, created_at, updated_at) VALUES (?, ?, NOW(), NOW())`
	_, err := db.Exec(query, name, description)
	return err
}

func UpdateRole(db *sqlx.DB, id int, name, description string) error {
	query := `UPDATE roles SET role_name = ?, description = ?, updated_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, name, description, id)
	return err
}

func DeleteRole(db *sqlx.DB, id int) error {
	query := `DELETE FROM roles WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
