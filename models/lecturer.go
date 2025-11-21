package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Lecturer struct {
	ID      int            `db:"id"`
	UserID  int            `db:"user_id"`
	NIP     sql.NullString `db:"nip"`
	NUPTK   sql.NullString `db:"nuptk"`
	Address sql.NullString `db:"address"`
	Phone   sql.NullString `db:"phone_number"`
}

type LecturerPosition struct {
	ID             int           `db:"id"`
	LecturerID     int           `db:"lecturer_id"`
	PositionID     int           `db:"position_id"`
	MajorID        sql.NullInt64 `db:"major_id"`
	StudyProgramID sql.NullInt64 `db:"study_program_id"`
}

type LecturerPositionDetail struct {
	PositionID   int
	PositionName string
	Scopetype    string
	ScopeName    sql.NullString
}

func GetLecturerPositionsByLecturerID(db *sqlx.DB, lecturerID int) ([]LecturerPositionDetail, error) {
	query := `SELECT 
    p.name AS position_name,
    CASE 
        WHEN lp.major_id IS NOT NULL THEN 'major'
        WHEN lp.study_program_id IS NOT NULL THEN 'prodi'
        ELSE 'none'
    END AS scope_type,
    COALESCE(m.name, sp.name) AS scope_name 
	FROM lecturer_positions lp 
	JOIN positions p ON lp.position_id = p.id 
	LEFT JOIN majors m ON lp.major_id = m.id 
	LEFT JOIN study_programs sp ON lp.study_program_id = sp.id 
	WHERE lp.lecturer_id = ?;`

	var positions []LecturerPositionDetail
	err := db.Select(&positions, query, lecturerID)
	if err != nil {
		return nil, err
	}
	return positions, nil
}
