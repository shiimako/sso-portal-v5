package models

import "github.com/jmoiron/sqlx"

type Position struct {
	ID            int    `db:"id"`
	Name string `db:"position_name"`
}

type DCPositionResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   []DCPosition `json:"data"`
}

type DCPosition struct {
	ID        int     `json:"id_jabatan"`
	Name      string  `json:"nama_jabatan"`
	DeletedAt *string `json:"deleted_at"`
}

func GetAllPositions(db *sqlx.DB) ([]Position, error) {
	var data []Position
	err := db.Select(&data, "SELECT id, position_name FROM positions ORDER BY position_name ASC")
	return data, err
}

func UpsertPosition(db *sqlx.DB, data DCPosition) error {
	query := `
		INSERT INTO positions 
			(id, position_name, created_at, updated_at, deleted_at)
		VALUES 
			(:id, :name, NOW(), NOW(), :deleted_at)
		ON DUPLICATE KEY UPDATE
			position_name = VALUES(position_name),
			updated_at = NOW(),
			deleted_at = VALUES(deleted_at);
	`
	params := map[string]interface{}{
		"id":         data.ID,
		"name":       data.Name,
		"deleted_at": data.DeletedAt,
	}
	_, err := db.NamedExec(query, params)
	return err
}