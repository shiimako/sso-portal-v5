package models

import "github.com/jmoiron/sqlx"

type Position struct {
	ID            int    `db:"id"`
	Name string `db:"position_name"`
}

func GetAllPositions(db *sqlx.DB) ([]Position, error) {
	var data []Position
	err := db.Select(&data, "SELECT id, position_name FROM positions ORDER BY position_name ASC")
	return data, err
}

func FindPositionByID(db *sqlx.DB, id int) (*Position, error) {
	var p Position
	query := "SELECT id, position_name FROM positions WHERE id = ? AND deleted_at IS NULL"
	err := db.Get(&p, query, id)
	return &p, err
}

func CreatePosition(db *sqlx.DB, name string) error {
	query := `INSERT INTO positions (position_name, created_at, updated_at) VALUES (?, NOW(), NOW())`
	_, err := db.Exec(query, name)
	return err
}

func UpdatePosition(db *sqlx.DB, id int, name string) error {
	query := `UPDATE positions SET position_name = ?, updated_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, name, id)
	return err
}

func DeletePosition(db *sqlx.DB, id int) error {
	query := `UPDATE positions SET deleted_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}