package models

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID           int            `db:"id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Status       string         `db:"status"`
	Avatar       sql.NullString `db:"avatar"`
	GoogleAvatar sql.NullString `db: "googleavatar"`
	Address      sql.NullString `db:"address"`
	Phone        sql.NullString `db:"address"`
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

type UserForm struct {
	ID       int
	Name     string
	Email    string
	RoleID   int
	RoleName string
	Status   string
	Address  *string
	Phone    *string

	NIM   *string
	NIP   *string
	NUPTK *string

	Positions []LecturerPosition
}

// =================
// READ & FIND FUNCTIONS
// =================
func FindUserByEmail(db *sqlx.DB, email string) (*FullUser, error) {

	query := `SELECT u.id, u.name, u.email, u.status, u.avatar, u.google_avatar, u.address, u.phone_number,
	r.id AS role_id, 
	r.role_name AS role_name 
	FROM users u 
	JOIN user_roles ur ON u.id = ur.user_id 
	JOIN roles r ON ur.role_id = r.id 
	WHERE LOWER(TRIM(u.email)) = LOWER(?)
	AND
	u.deleted_at IS NULL 
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
	u.deleted_at IS NULL
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
        WHERE u.deleted_at IS NULL
    `

	args := []interface{}{}

	if search != "" {
		query += " AND (u.name LIKE ? OR u.email LIKE ?) "
		args = append(args, "%"+search+"%", "%"+search+"%")
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

func GetContact(db *sqlx.DB, email string) (*AdminContact, error) {

	var contact AdminContact

	query := `SELECT email, phone_number FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(
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

// =================
// CREATE FUNCTIONS
// =================
func CreateUser(db *sqlx.DB, form UserForm) (int64, error) {

	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `INSERT INTO users (name, email, status, address, phone_number) VALUES (?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, form.Name, form.Email, form.Status, form.Address, form.Phone)
	if err != nil {
		return 0, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if form.RoleID != 0 {
		_, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, form.RoleID)
		if err != nil {
			return 0, err
		}
	}

	if form.RoleName == "mahasiswa" {
		query = `INSERT INTO students (user_id, nim) VALUES (?, ?)`
		_, err := tx.Exec(query, userID, form.NIM)
		if err != nil {
			return 0, err
		}
	} else if form.RoleName == "dosen" {
		query = `INSERT INTO lecturers (user_id, nip, nuptk) VALUES (?, ?, ?)`
		res2, err := tx.Exec(query, userID, form.NIP, form.NUPTK)
		if err != nil {
			return 0, err
		}
		lecturerID, err := res2.LastInsertId()
		if err != nil {
			return 0, err
		}
		for _, pos := range form.Positions {
			_, err = tx.Exec(`INSERT INTO lecturer_positions (lecturer_id, position_id, major_id, study_program_id, start_date, end_date) 
			VALUES (?, ?, ?, ?, ?, ?)`,
				lecturerID, pos.PositionID, pos.MajorID, pos.StudyProgramID, pos.StartDate, pos.EndDate)
			if err != nil {
				return 0, err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// =================
// UPDATE FUNCTIONS
// =================

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

func UpdateUser(db *sqlx.DB, form UserForm) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	_, err = tx.Exec(`UPDATE users SET name = ?, email = ?, status = ?, address = ?, phone_number = ?, updated_at = NOW() WHERE id = ?`,
		form.Name, form.Email, form.Status, form.Address, form.Phone, form.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id=?`, form.ID)
	_, err = tx.Exec(`DELETE FROM students WHERE user_id=?`, form.ID)
	_, err = tx.Exec(`DELETE FROM lecturer_positions WHERE lecturer_id IN (SELECT id FROM lecturers WHERE user_id=?)`, form.ID)
	_, err = tx.Exec(`DELETE FROM lecturers WHERE user_id=?`, form.ID)

	if err != nil { return err }

	if form.RoleID != 0 {
		_, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, form.ID, form.RoleID)
		if err != nil { return err }
	}

	if form.RoleName == "mahasiswa" {
		_, err = tx.Exec(`INSERT INTO students (user_id, nim) VALUES (?, ?)`, form.ID, form.NIM)
		if err != nil { return err }
	} else if form.RoleName == "dosen" {
		res, err := tx.Exec(`INSERT INTO lecturers (user_id, nip, nuptk) VALUES (?, ?, ?)`, form.ID, form.NIP, form.NUPTK)
		if err != nil { return err }
		
		lecturerID, _ := res.LastInsertId()
		
		for _, pos := range form.Positions {
			_, err = tx.Exec(`INSERT INTO lecturer_positions (lecturer_id, position_id, major_id, study_program_id, start_date, end_date) 
			VALUES (?, ?, ?, ?, ?, ?)`,
				lecturerID, pos.PositionID, pos.MajorID, pos.StudyProgramID, pos.StartDate, pos.EndDate)
			if err != nil { return err }
		}
	}

	return tx.Commit()
}

// =================
// DELETE FUNCTIONS
// =================
func DeleteUser(db *sqlx.DB, userID int) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	_, err = tx.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM lecturers WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM students WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

// ==========================================
// HELPER: PARSE JSON POSITIONS
// ==========================================
func ParseLecturerPositions(jsonStr string) ([]LecturerPosition, error) {
	if jsonStr == "" || jsonStr == "[]" {
		return nil, nil
	}

	var raw []PositionFormJSON
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, err
	}

	var result []LecturerPosition
	for _, p := range raw {
        if p.PositionID == 0 { continue }

		pos := LecturerPosition{
			PositionID: p.PositionID,
			StartDate:  GetPtr(p.StartDate),
			EndDate:    GetPtr(p.EndDate), 
		}

		if p.Scope == "major" && p.MajorID > 0 {
			pos.MajorID = sql.NullInt64{Int64: int64(p.MajorID), Valid: true}
		} else if p.Scope == "prodi" && p.ProdiID > 0 {
			pos.StudyProgramID = sql.NullInt64{Int64: int64(p.ProdiID), Valid: true}
		}

		result = append(result, pos)
	}
	return result, nil
}

// Helper GetPtr
func GetPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
