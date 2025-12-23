package models

import "github.com/jmoiron/sqlx"

type StudyProgram struct {
	ID        int    `db:"id"`
	Name      string `db:"study_program_name"`
	MajorName string `db:"major_name"`
	MajorID   int    `db:"major_id"`
}

func GetAllStudyPrograms(db *sqlx.DB) ([]StudyProgram, error) {
	var data []StudyProgram
	query := `
		SELECT sp.id, sp.study_program_name, sp.major_id, m.major_name 
		FROM study_programs sp
		LEFT JOIN majors m ON sp.major_id = m.id
		WHERE sp.deleted_at IS NULL
		ORDER BY m.major_name ASC, sp.study_program_name ASC
	`
	err := db.Select(&data, query)
	return data, err
}

func FindStudyProgramByID(db *sqlx.DB, id int) (*StudyProgram, error) {
	var sp StudyProgram
	query := `
		SELECT sp.id, sp.study_program_name, sp.major_id, m.major_name 
		FROM study_programs sp
		LEFT JOIN majors m ON sp.major_id = m.id
		WHERE sp.id = ? AND sp.deleted_at IS NULL
	`
	err := db.Get(&sp, query, id)
	return &sp, err
}

func CreateStudyProgram(db *sqlx.DB, name string, majorID int) error {
	query := `INSERT INTO study_programs (study_program_name, major_id, created_at, updated_at) VALUES (?, ?, NOW(), NOW())`
	_, err := db.Exec(query, name, majorID)
	return err
}

func UpdateStudyProgram(db *sqlx.DB, id int, name string, majorID int) error {
	query := `UPDATE study_programs SET study_program_name = ?, major_id = ?, updated_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, name, majorID, id)
	return err
}

func DeleteStudyProgram(db *sqlx.DB, id int) error {
	query := `UPDATE study_programs SET deleted_at = NOW() WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}