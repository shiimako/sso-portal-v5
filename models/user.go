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
	GoogleAvatar sql.NullString `db: "googleavatar"`
	Address sql.NullString `db:"address"`
	Phone sql.NullString `db:"address"`
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

type AdminContact struct {
	Phone string 
	Email string
}

func FindUserByEmail(db *sqlx.DB, email string) (*FullUser, error) {

	query := `SELECT u.id, u.name, u.email, u.status, u.avatar, u.google_avatar, u.address, u.phone_number,
	r.id AS role_id, 
	r.role_name AS role_name 
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
		&fu.GoogleAvatar,
		&fu.Address,
		&fu.Phone,
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
		err := db.Get(&s, "SELECT id, user_id, nim FROM students WHERE user_id = ?", fu.ID)
		if err == nil {
			fu.Student = &s
		}
	} else if role.Name == "dosen" {
		var l Lecturer
		err := db.Get(&l, "SELECT id, user_id, nip, nuptk FROM lecturers WHERE user_id = ?", fu.ID)
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

	query := `SELECT u.id, u.name, u.email, u.status, u.avatar, u.google_avatar, u.address, u.phone_number,
	r.id AS role_id, 
	r.role_name AS role_name 
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
		&fu.GoogleAvatar,
		&fu.Address,
		&fu.Phone,
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
		err := db.Get(&s, "SELECT id, user_id, nim FROM students WHERE user_id = ?", fu.ID)
		if err == nil {
			fu.Student = &s
		}
	} else if role.Name == "dosen" {
		var l Lecturer
		err := db.Get(&l, "SELECT id, user_id, nip, nuptk FROM lecturers WHERE user_id = ?", fu.ID)
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
        UPDATE users SET avatar = ?, google_avatar = ?, updated_at = NOW() WHERE id = ?
    `, avatarURL, avatarURL, userID)

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

func GetAllUsers(db *sqlx.DB, page, pagesize int, search, role string) ([]UserListItem, error) {

    offset := (page - 1) * pagesize

    query := `
        SELECT 
            u.id, 
            u.name, 
            u.email, 
            u.status, 
            r.role_name AS role
        FROM users u
        JOIN user_roles ur ON ur.user_id = u.id
        JOIN roles r ON r.id = ur.role_id
        WHERE 1 = 1
    `

    args := []interface{}{}

    if search != "" {
        query += " AND u.name LIKE ? "
        args = append(args, "%"+search+"%")
    }

    if role != "" {
        query += " AND r.role_name = ? "
        args = append(args, role)
    }

    query += `
        ORDER BY u.id ASC
        LIMIT ? OFFSET ?
    `
    args = append(args, pagesize, offset)

    var users []UserListItem
    err := db.Select(&users, query, args...)
    if err != nil {
        return nil, err
    }

    return users, nil
}

func UpdateUserProfile(db *sqlx.DB, userID int, address, phone, avatarPath string) error {
	query := `UPDATE users SET address = ?, phone_number = ?, updated_at = NOW()`
	args := []interface{}{address, phone}

	if avatarPath != "" {
		query += `, avatar = ?`
		args = append(args, avatarPath)
	}

	query += ` WHERE id = ?`
	args = append(args, userID)

	_, err := db.Exec(query, args...)
	return err
}

func GetAdminContact(db *sqlx.DB) (*AdminContact, error){

	var contact AdminContact

	query := `SELECT email, phone_number FROM users WHERE id = 1`
	err := db.QueryRow(query).Scan(
		&contact.Email,
		&contact.Phone,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &contact, nil
}

