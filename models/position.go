package models

import "github.com/jmoiron/sqlx"

type Position struct {
	ID            int    `db:"id"`
	Position_Name string `db:"position_name"`
}

func GetAllPosition(db *sqlx.DB) ([]Position, error){
	query := `SELECT id, position_name FROM positions`

	var pos []Position

	err := db.Select(&pos, query)
	return pos, err
}