package models

import "github.com/jmoiron/sqlx"

// ==============================
// LOCAL STRUCT 
// ==============================
type Major struct {
	ID         int    `db:"id"`
	Name string `db:"major_name"`
}

// ==============================
// DATA CENTER STRUCT 
// ==============================
type DCMajorResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   []DCMajor `json:"data"`
}
type DCMajor struct {
	ID   int    `json:"id_jurusan"`
	Name string `json:"nama_jurusan"`
	DeletedAt *string `json:"deleted_at"`
}

func GetAllMajors(db *sqlx.DB) ([]Major, error) {
	var data []Major
	query := "SELECT id, major_name FROM majors WHERE deleted_at IS NULL ORDER BY major_name ASC"
	err := db.Select(&data, query)
	return data, err
}

func UpsertMajor(db *sqlx.DB, data DCMajor) error {
	query := `
		INSERT INTO majors (id, major_name, created_at, updated_at, deleted_at)
		VALUES (:id, :name, NOW(), NOW(), :deleted_at)
		ON DUPLICATE KEY UPDATE
			major_name = VALUES(major_name),
			updated_at = NOW(),
			deleted_at = VALUES(deleted_at);
	`
	params := map[string]interface{}{
		"id":   data.ID,
		"name": data.Name,
		"deleted_at": data.DeletedAt,
	}
	_, err := db.NamedExec(query, params)
	return err
}