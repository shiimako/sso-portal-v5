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
}

type LecturerPosition struct {
	ID             int           `db:"id"`
	LecturerID     int           `db:"lecturer_id"`
	PositionID     int           `db:"position_id"`
	PositionName   string        `db:"position_name"` // Join result
	MajorID        sql.NullInt64 `db:"major_id"`
	StudyProgramID sql.NullInt64 `db:"study_program_id"`
	StartDate      *string       `db:"start_date"` // Pakai Pointer (*)
	EndDate        *string       `db:"end_date"`   // Pakai Pointer (*)
    
    // Field Tambahan untuk Display di Frontend (Join Result)
	Scopetype string         `db:"scopetype"` // 'major', 'prodi', 'none'
	ScopeName sql.NullString `db:"scopename"`
}

type PositionFormJSON struct {
	PositionID int    `json:"position_id"`
	Scope      string `json:"scope"` // major, prodi, none
	MajorID    int    `json:"major_id"`
	ProdiID    int    `json:"prodi_id"`
	StartDate  string `json:"start_date"` 
	EndDate    string `json:"end_date"`
}

type LecturerPositionDetail struct {
	PositionID   int
	PositionName string
	Scopetype    string
	ScopeName    sql.NullString
}

func GetLecturerPositionsByLecturerID(db *sqlx.DB, lecturerID int) ([]LecturerPositionDetail, error) {
	query := `SELECT 
    p.position_name AS positionname,
    CASE 
        WHEN lp.major_id IS NOT NULL THEN 'major'
        WHEN lp.study_program_id IS NOT NULL THEN 'prodi'
        ELSE 'none'
    END AS scopetype,
    COALESCE(m.major_name, sp.study_program_name) AS scopename 
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

