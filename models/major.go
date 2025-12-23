package models

import "github.com/jmoiron/sqlx"

type Major struct {
	ID         int    `db:"id"`
	Name string `db:"major_name"`
}


func GetAllMajors(db *sqlx.DB) ([]Major, error) {
	var data []Major
	query := "SELECT id, major_name FROM majors WHERE deleted_at IS NULL ORDER BY major_name ASC"
	err := db.Select(&data, query)
	return data, err
}

func FindMajorByID(db *sqlx.DB, id int) (*Major, error) {
	var m Major
	query := "SELECT id, major_name FROM majors WHERE id = ? AND deleted_at IS NULL"
	err := db.Get(&m, query, id)
	return &m, err
}

func CreateMajor(db *sqlx.DB, name string) error {
	query := `INSERT INTO majors (major_name, created_at, updated_at) VALUES (?, NOW(), NOW())`
	_, err := db.Exec(query, name)
	return err
}

func UpdateMajor(db *sqlx.DB, id int, name string) error {
	query := `UPDATE majors SET major_name = ?, updated_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, name, id)
	return err
}

func DeleteMajor(db *sqlx.DB, id int) error {
	query := `UPDATE majors SET deleted_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}