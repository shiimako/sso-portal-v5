package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Category struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Sort int    `db:"sort_order"`
}

func GetAllCategories(db *sqlx.DB) ([]Category, error) {
	var categs []Category
	err := db.Select(&categs, "SELECT * FROM categories ORDER BY sort_order ASC")
	return categs, err
}

func ListCategories(db *sqlx.DB, page int, pagesize int) ([]Category, error) {
	offset := (page - 1) * pagesize

	var cats []Category
	query := `SELECT * FROM categories`
	params := []interface{}{}

	query += ` ORDER BY id ASC LIMIT ? OFFSET ? `
	params = append(params, pagesize, offset)

	err := db.Select(&cats, query, params...)
	if err != nil {
		return nil, err
	}
	return cats, nil
}

func FindCategoryByID(db *sqlx.DB, id int) (Category, error) {
	var cats Category
	query := `SELECT * FROM categories WHERE id = ?`
	err := db.Get(&cats, query, id)
	if err != nil {
		return cats, err
	}
	return cats, nil
}

func CreateCategory(db *sqlx.DB, name string, sort int) error {
	query := `INSERT INTO categories (name, sort_order) VALUES (?, ?)`
	_, err := db.Exec(query, name, sort)
	return err
}

func UpdateCategory(db *sqlx.DB, id string, name string, sort int) error {
	query := `UPDATE categories SET name = ?, sort_order = ? WHERE id = ?`
	result, err := db.Exec(query, name, sort, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func DeleteCategory(db *sqlx.DB, id int) error {
	query := `DELETE FROM categories WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func GetMaxSort(db *sqlx.DB) (int, error) {
	var max int
	query := `SELECT COALESCE(MAX(sort_order), 0) FROM categories`

	err := db.Get(&max, query)
	return max, err

}

func IsSortExists(db *sqlx.DB, sort int) (bool, error) {
	var exists bool

	query := `SELECT EXISTS (SELECT 1 FROM categories WHERE sort_order = ?)`

	err := db.Get(&exists, query, sort)
	if err != nil {
		return false, err
	}

	return exists, nil
}
