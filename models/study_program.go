package models

import "github.com/jmoiron/sqlx"

type StudyProgram struct {
	ID        int    `db:"id"`
	Name      string `db:"study_program_name"`
	MajorName string `db:"major_name"`
	MajorID   int    `db:"major_id"`
}

type DCStudyProgram struct {
	ID        int     `json:"id_prodi"`
	Name      string  `json:"nama_prodi"`
	MajorID   int     `json:"id_jurusan"`
	DeletedAt *string `json:"deleted_at"`
}

type DCStudyProgramResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   []DCStudyProgram `json:"data"`
}

func GetAllStudyPrograms(db *sqlx.DB) ([]StudyProgram, error) {
	var data []StudyProgram
	query := `
		SELECT sp.id, sp.study_program_name, sp.major_id, m.major_name 
		FROM study_programs sp
		LEFT JOIN majors m ON sp.major_id = m.id
		ORDER BY m.major_name ASC, sp.study_program_name ASC
	`
	err := db.Select(&data, query)
	return data, err
}

func UpsertStudyPrograms(db *sqlx.DB, data DCStudyProgram) error {
	query := `
		INSERT INTO study_programs 
			(id, study_program_name, major_id, created_at, updated_at, deleted_at)
		VALUES 
			(:id, :name, :major_id, NOW(), NOW(), :deleted_at)
		ON DUPLICATE KEY UPDATE
			study_program_name = VALUES(study_program_name),
			major_id = VALUES(major_id),
			updated_at = NOW(),
			deleted_at = VALUES(deleted_at);
	`
	params := map[string]interface{}{
		"id":         data.ID,
		"name":       data.Name,
		"major_id":   data.MajorID, 
		"deleted_at": data.DeletedAt,
	}
	_, err := db.NamedExec(query, params)
	return err
}