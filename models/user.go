package models

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID     int            `db:"id"`
	Name   string         `db:"name"`
	Email  string         `db:"email"`
	Status string         `db:"status"`
	Avatar sql.NullString `db:"avatar"`
}
type UserRole struct {
	RoleID int    `db:"role_id"`
	Name   string `db:"role_name"`
}
type FullUser struct {
	User
	Roles     []UserRole
	Student   *Student
	Lecturer  *Lecturer
	Positions []LecturerPosition
}
type UserListItem struct {
	ID     int            `db:"id"`
	Name   string         `db:"name"`
	Email  string         `db:"email"`
	Status string         `db:"status"`
	Avatar sql.NullString `db:"avatar"`
	Role   string         `db:"role"`
}

func FindUserByEmail(db *sqlx.DB, email string) (*FullUser, error) {

	query := `SELECT u.id, u.name, u.email, u.status, u.avatar, 
	r.id AS role_id, 
	r.name AS role_name 
	FROM users u 
	JOIN user_roles ur ON u.id = ur.user_id 
	JOIN roles r ON ur.role_id = r.id 
	WHERE LOWER(TRIM(u.email)) = LOWER(?)
	AND
	deleted_at IS NULL 
	LIMIT 1;`

	var fu FullUser
	var role UserRole
	err := db.QueryRow(query, email).Scan(
		&fu.ID,
		&fu.Name,
		&fu.Email,
		&fu.Status,
		&fu.Avatar,
		&role.RoleID,
		&role.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	fu.Roles = append(fu.Roles, role)

	// Ambil data tambahan berdasarkan peran
	if role.Name == "mahasiswa" {
		var s Student
		err := db.Get(&s, "SELECT id, user_id, nim, address, phone_number FROM students WHERE user_id = ?", fu.ID)
		if err == nil {
			fu.Student = &s
		}
	} else if role.Name == "dosen" {
		var l Lecturer
		err := db.Get(&l, "SELECT id, user_id, nip, nuptk, address, phone_number FROM lecturers WHERE user_id = ?", fu.ID)
		if err == nil {
			fu.Lecturer = &l
		}

		// Ambil posisi dosen jika ada
		var positions []LecturerPosition
		err = db.Select(&positions, "SELECT id, lecturer_id, position_id, major_id, study_program_id FROM lecturer_positions WHERE lecturer_id = ?", fu.Lecturer.ID)
		if err == nil {
			fu.Positions = positions
		}
	}

	return &fu, nil
}

func FindUserByID(db *sqlx.DB, id int) (*FullUser, error) {

	query := `SELECT u.id, u.name, u.email, u.status, u.avatar, 
	r.id AS role_id, 
	r.name AS role_name 
	FROM users u 
	JOIN user_roles ur ON u.id = ur.user_id 
	JOIN roles r ON ur.role_id = r.id 
	WHERE u.id = ? 
	AND
	deleted_at IS NULL
	LIMIT 1;`

	var fu FullUser
	var role UserRole
	err := db.QueryRow(query, id).Scan(
		&fu.ID,
		&fu.Name,
		&fu.Email,
		&fu.Status,
		&fu.Avatar,
		&role.RoleID,
		&role.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	fu.Roles = append(fu.Roles, role)

	// Ambil data tambahan berdasarkan peran
	if role.Name == "mahasiswa" {
		var s Student
		err := db.Get(&s, "SELECT id, user_id, nim, address, phone_number FROM students WHERE user_id = ?", fu.ID)
		if err == nil {
			fu.Student = &s
		}
	} else if role.Name == "dosen" {
		var l Lecturer
		err := db.Get(&l, "SELECT id, user_id, nip, nuptk, address, phone_number FROM lecturers WHERE user_id = ?", fu.ID)
		if err == nil {
			fu.Lecturer = &l
		}

		var positions []LecturerPosition
		err = db.Select(&positions, "SELECT id, lecturer_id, position_id, major_id, study_program_id FROM lecturer_positions WHERE lecturer_id = ?", fu.Lecturer.ID)
		if err == nil {
			fu.Positions = positions
		}
	}

	return &fu, nil
}

func UpdateUserAvatar(db *sqlx.DB, userID int, avatarURL string) error {
	res, err := db.Exec(`
        UPDATE users SET avatar = ?, updated_at = NOW() WHERE id = ?
    `, avatarURL, userID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user id %d tidak ditemukan", userID)
	}

	return nil
}

func GetAllUsers(db *sqlx.DB) ([]UserListItem, error) {
	query := `SELECT u.id, u.name, u.email, u.status, r.name AS role 
	FROM users u 
	JOIN user_roles ur ON ur.user_id = u.id 
	JOIN roles r ON r.id = ur.role_id 
	ORDER BY u.id DESC;`
	
	var users []UserListItem
	err := db.Select(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}
