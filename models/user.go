package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

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

type DCUserResponse struct {
	Code   int      `json:"code"`
	Status string   `json:"status"`
	Data   []DCUser `json:"data"`
}

type DCUser struct {
	ID        int     `json:"id_user"`
	Name      string  `json:"nama_lengkap"`
	Email     string  `json:"email"`
	Status    string  `json:"status"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at"`

	// Handle Role: JSON User 1 pakai Object, User 2 & 3 pakai Integer langsung
	// Kita sediakan dua-duanya, nanti di logic kita cek mana yang isi
	RoleObj *struct {
		ID   int    `json:"id_role"`
		Nama string `json:"nama"`
	} `json:"role"`
	RoleID *int `json:"id_role"`

	// Profile
	Profile struct {
		Address string `json:"alamat"` // JSON User 2 pakai "id_alamat" (typo di JSON?), kita coba handle
		AddressAlt string `json:"id_alamat"` // Jaga-jaga kalau key-nya berubah
		Phone   string `json:"no_hp"`
	} `json:"profile"`

	// Academic (Polimorfik: Bisa Student, Bisa Dosen, Bisa Kosong)
	Academic DCAcademic `json:"academic"`
}

type DCAcademic struct {
	// Fields Mahasiswa
	NIM *string `json:"nim"`
	
	// Fields Dosen
	NIP   *string `json:"nip"` // Kadang ada kadang ngga
	NUPTK *string `json:"nuptk"`
	
	// Jabatan Dosen (Array)
	Positions []DCLecturerPos `json:"jabatan_dosen"`
}

type DCLecturerPos struct {
	IDJabatan  int     `json:"id_jabatan"`
	IDJurusan  *int    `json:"id_jurusan"` // Pointer biar bisa null
	IDProdi    *int    `json:"id_prodi"`
	StartDate  string  `json:"tanggal_mulai"`   // Format: "16-02-2021"
	EndDate    *string `json:"tanggal_selesai"` // Format: "16-02-2025"
	Status     string  `json:"status"`
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

func UpsertFullUser(db *sqlx.DB, u DCUser) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// A. NORMALISASI DATA SEBELUM INSERT
	
	// 1. Tentukan Role ID (Prioritas Object, lalu Int)
	finalRoleID := 0
	if u.RoleObj != nil {
		finalRoleID = u.RoleObj.ID
	} else if u.RoleID != nil {
		finalRoleID = *u.RoleID
	}

	// 2. Tentukan Alamat (Handle typo JSON)
	finalAddress := u.Profile.Address
	if finalAddress == "" {
		finalAddress = u.Profile.AddressAlt
	}

	// 3. Lowercase Status (Jaga-jaga enum DB lowercase)
	finalStatus := strings.ToLower(u.Status)

	// 4. Parsing DeletedAt
	var dbDeletedAt *string
	if u.DeletedAt != nil {
		t, _ := time.Parse(time.RFC3339, *u.DeletedAt)
		str := t.Format("2006-01-02 15:04:05")
		dbDeletedAt = &str
	}

	// -------------------------------------------------
	// B. QUERY 1: TABEL USERS
	// -------------------------------------------------
	qUser := `
		INSERT INTO users (id, name, email, status, address, phone_number, created_at, updated_at, deleted_at)
		VALUES (:id, :name, :email, :status, :address, :phone, NOW(), NOW(), :deleted_at)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			email = VALUES(email),
			status = VALUES(status),
			address = VALUES(address),
			phone_number = VALUES(phone_number),
			updated_at = NOW(),
			deleted_at = VALUES(deleted_at);
	`
	userParams := map[string]interface{}{
		"id": u.ID, "name": u.Name, "email": u.Email, "status": finalStatus,
		"address": finalAddress, "phone": u.Profile.Phone, "deleted_at": dbDeletedAt,
	}
	if _, err := tx.NamedExec(qUser, userParams); err != nil {
		return fmt.Errorf("user error: %v", err)
	}

	// -------------------------------------------------
	// C. QUERY 2: TABEL USER_ROLES (Reset Role)
	// -------------------------------------------------
	if _, err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", u.ID); err != nil {
		return fmt.Errorf("delete role error: %v", err)
	}
	if finalRoleID != 0 {
		if _, err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", u.ID, finalRoleID); err != nil {
			return fmt.Errorf("insert role error: %v", err)
		}
	}

	// -------------------------------------------------
	// D. QUERY 3: DETAIL (MAHASISWA / DOSEN)
	// -------------------------------------------------
	
	// Cek Role ID: 1=Admin, 2=Dosen, 3=Mahasiswa (Sesuaikan dengan DB kamu)
	// Asumsi berdasarkan JSON kamu: 2=Mahasiswa?, 3=Dosen? 
	// (User 2 punya NIM -> Mahasiswa. User 3 punya NUPTK -> Dosen)
	
	// KASUS MAHASISWA (Punya NIM)
	if u.Academic.NIM != nil {
		qMhs := `INSERT INTO students (user_id, nim) VALUES (?, ?) ON DUPLICATE KEY UPDATE nim = VALUES(nim)`
		if _, err := tx.Exec(qMhs, u.ID, *u.Academic.NIM); err != nil {
			return fmt.Errorf("student error: %v", err)
		}
	}

	// KASUS DOSEN (Punya NUPTK atau Jabatan)
	if u.Academic.NUPTK != nil || len(u.Academic.Positions) > 0 {
		// 1. Upsert Tabel Lecturers
		qDosen := `INSERT INTO lecturers (user_id, nip, nuptk) VALUES (?, ?, ?) 
				   ON DUPLICATE KEY UPDATE nip = VALUES(nip), nuptk = VALUES(nuptk)`
		if _, err := tx.Exec(qDosen, u.ID, u.Academic.NIP, u.Academic.NUPTK); err != nil {
			return fmt.Errorf("lecturer error: %v", err)
		}

		// 2. Ambil ID Lecturer (Bukan User ID)
		var lecturerID int
		if err := tx.Get(&lecturerID, "SELECT id FROM lecturers WHERE user_id = ?", u.ID); err != nil {
			return fmt.Errorf("get lecturer id error: %v", err)
		}

		// 3. Sync Jabatan Dosen (Hapus lama, Insert baru)
		if _, err := tx.Exec("DELETE FROM lecturer_positions WHERE lecturer_id = ?", lecturerID); err != nil {
			return fmt.Errorf("delete positions error: %v", err)
		}

		for _, pos := range u.Academic.Positions {
			parsedStart, _ := time.Parse("02-01-2006", pos.StartDate)
			startDateDB := parsedStart.Format("2006-01-02")

			var endDateDB *string
			if pos.EndDate != nil && *pos.EndDate != "" {
				parsedEnd, _ := time.Parse("02-01-2006", *pos.EndDate)
				str := parsedEnd.Format("2006-01-02")
				endDateDB = &str
			}

			qPos := `
                INSERT INTO lecturer_positions 
                (lecturer_id, position_id, major_id, study_program_id, start_date, end_date, created_at, updated_at)
			    VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
            `
			
			// CreatedAt pakai tanggal mulai jabatan aja biar relevan
			if _, err := tx.Exec(qPos, lecturerID, pos.IDJabatan, pos.IDJurusan, pos.IDProdi, startDateDB, endDateDB); err != nil {
				return fmt.Errorf("insert position error: %v", err)
			}
		}
	}

	return tx.Commit()
}

