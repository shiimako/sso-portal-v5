package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

// UserWithRoles adalah struct untuk menampilkan daftar pengguna di tabel admin.
type UserWithRoles struct {
	ID     int `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Status string `db:"status"`
	Roles  string `db:"roles"`
}

// User adalah struct untuk data user lengkap.
type User struct {
	ID         int  `db:"id"`
	Name       string `db:"name"`
	Email      string `db:"email"`
	Status     string `db:"status"`
	Avatar 		sql.NullString `db:"avatar"`

	Roles      []int          
	RolesStr   string         
	Attributes map[int]string 

	Student_ID sql.NullInt64  `db:"student_id"`
	NIM  sql.NullString `db:"nim"`

	Lecturer_ID sql.NullInt64  `db:"lecturer_id"`
	NIP  sql.NullString `db:"nip"`
	NIDN sql.NullString `db:"nidn"`

	Address sql.NullString `db:"address"`
	Phone   sql.NullString `db:"phone_number"`
}

// Attribute adalah struct untuk data atribut.
type Attribute struct {
	ID   int
	Name string
}

// GetAllUsers mengambil semua data pengguna beserta perannya dari database.
func GetAllUsers(db *sqlx.DB) ([]UserWithRoles, error) {
	var users []UserWithRoles

	query := `
        SELECT u.id, u.name, u.email, u.status, GROUP_CONCAT(r.name SEPARATOR ', ') as roles
        FROM users u
        LEFT JOIN user_roles ur ON u.id = ur.user_id
        LEFT JOIN roles r ON ur.role_id = r.id
        GROUP BY u.id
        ORDER BY u.id ASC`

	err := db.Select(&users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetAllRoles mengambil semua peran dasar (selain admin) dari database.
func GetAllRoles(db *sqlx.DB) ([]Role, error) {

	var roles []Role

	// Ambil semua peran dasar (selain admin) dari database untuk ditampilkan di dropdown
	query := `SELECT id, name FROM roles WHERE type = 'base' AND name != 'admin' ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// FindUserByID sekarang hanya fokus mengambil data satu user.
func FindUserByID(db *sqlx.DB, id string) (User, error) {
	var user User
	err := db.Get(&user, `SELECT id, name, email, status, avatar FROM users WHERE id=?`, id)
	if err != nil {
		return user, err
	}

	roles, err := GetUserRolesAndAttributes(db, user.ID)
	if err != nil {
		return user, err
	}

	// 3. Proses peran dan cari tahu apakah dia dosen atau mahasiswa
	var hasStudentRole, hasLecturerRole bool
	user.Attributes = make(map[int]string)
	var sb strings.Builder
	sb.WriteString(",")

	for _, r := range roles {
		// Asumsi 'base' roles punya ID 2 (dosen) dan 3 (mahasiswa)
		if r.Type == "base" {
			roleID := 0 // Cari ID berdasarkan nama r.Name jika perlu
			if r.Name == "dosen" { 
				roleID = 2 // Ganti hardcode jika perlu
				hasLecturerRole = true
			}
			if r.Name == "mahasiswa" { 
				roleID = 3 // Ganti hardcode jika perlu
				hasStudentRole = true
			}
			
			if roleID != 0 {
				user.Roles = append(user.Roles, roleID)
				sb.WriteString(fmt.Sprintf("%d,", roleID))
			}
		}
	}
	user.RolesStr = sb.String()


	// 4. Ambil data profil (NIM/NIP/Alamat) berdasarkan perannya
	if hasStudentRole {
		// Ambil data mahasiswa
		err = db.QueryRowx(`SELECT id AS student_id, nim, address, phone_number FROM students WHERE user_id=?`, id).StructScan(&user)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Warning: Gagal ambil detail mahasiswa: %v", err)
		}
	} else if hasLecturerRole {
		// Ambil data dosen
		err = db.QueryRowx(`SELECT id as lecturer_id, nip, nidn, address, phone_number FROM lecturers WHERE user_id=?`, id).StructScan(&user)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Warning: Gagal ambil detail dosen: %v", err)
		}
	}

	return user, nil
}

// Fungsi ini khusus untuk mengambil semua kemungkinan atribut.
func GetAllAttributes(db *sqlx.DB) ([]Attribute, error) {
	var attributes []Attribute
	err := db.Select(&attributes, `SELECT id, name FROM roles WHERE type = 'attribute' ORDER BY name`)
	return attributes, err
}

// Struct untuk data user yang diambil dari database
func StoreNewUser(db *sqlx.DB, name, email, roleID, nip, nidn, nim, address, phone string) error {
	// Mulai transaksi
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Defer a rollback in case of panic
	defer tx.Rollback()

	// Simpan data ke tabel users
	query1 := `INSERT INTO users (name, email, status) VALUES (?, ?, 'aktif')`
	result1, err := tx.Exec(query1, name, email)
	if err != nil {
		return err
	}

	// Dapatkan ID user baru (auto-increment)
	userID, err := result1.LastInsertId()
	if err != nil {
		return err
	}

	// Simpan data ke tabel user_roles
	query2 := `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`
	_, err = tx.Exec(query2, userID, roleID)
	if err != nil {
		return err
	}

	// Simpan data ke tabel dosen atau mahasiswa sesuai peran
	switch roleID {
	case "2": // dosen
		query3 := `INSERT INTO lecturers (user_id, nip, nidn, address, phone_number) VALUES (?, ?, ?, ?, ?)`
		_, err = tx.Exec(query3, userID, nip, nidn, address, phone)
	case "3": // mahasiswa
		query3 := `INSERT INTO students (user_id, nim, address, phone_number) VALUES (?, ?, ?, ?)`
		_, err = tx.Exec(query3, userID, nim, address, phone)
	}

	if err != nil {
		return err
	}

	// Commit transaksi
	return tx.Commit()
}

// UpdateUser memperbarui data user di database.
func UpdateUser(db *sqlx.DB, id, name, email, status string, baseRoleIDs []string, attributeIDs []string, attributeScopes map[string]string, nim, nip, nidn string) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE users SET name=?, email=?, status=? WHERE id=?`, name, email, status, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id=? AND role_id IN (SELECT id FROM roles WHERE type='base')`, id)
	if err != nil {
		return err
	}

	for _, rid := range baseRoleIDs {
		_, err := tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, id, rid)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id=? AND role_id IN (SELECT id FROM roles WHERE type='attribute')`, id)
	if err != nil {
		return err
	}

	if len(attributeIDs) > 0 {
		stmt, err := tx.Preparex(`INSERT INTO user_roles (user_id, role_id, scope) VALUES (?, ?, ?)`)
		if err != nil {
			return err
		}
		for _, attrID := range attributeIDs {
			scope := attributeScopes[attrID]
			_, err := stmt.Exec(id, attrID, scope)
			if err != nil {
				return err
			}
		}
	}

	if contains(baseRoleIDs, "3") { // mahasiswa
		if nim != "" {
			_, err = tx.Exec(`INSERT INTO students (user_id, nim) VALUES (?, ?) ON DUPLICATE KEY UPDATE nim=?`, id, nim, nim)
			if err != nil {
				return err
			}
		}
	}
	if contains(baseRoleIDs, "2") { // dosen
		if nip != "" || nidn != "" {
			_, err = tx.Exec(`INSERT INTO lecturers (user_id, nip, nidn) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE nip=?, nidn=?`, id, nip, nidn, nip, nidn)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// DeleteUser menghapus user dari database berdasarkan ID.
func DeleteUser(db *sqlx.DB, id string) error {
	// Mulai transaksi
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Defer a rollback in case of panic
	defer tx.Rollback()

	// Hapus dari tabel user_roles
	_, err = tx.Exec(`DELETE FROM users WHERE id=?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Helper cek role_id ada di slice
func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

// UpdateUserAvatar HANYA mengupdate avatar
func UpdateUserAvatar(db *sqlx.DB, userID int, avatarURL string) error {
	_, err := db.Exec(`UPDATE users SET avatar = ? WHERE id = ?`, avatarURL, userID)
	return err
}

// UpdateUserProfile mengupdate data yang bisa diubah oleh user sendiri
func UpdateUserProfile(db *sqlx.DB, userID int, name, address, phone, avatarPath string) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update tabel users
	if avatarPath != "" {
		// Jika ada avatar baru, update nama DAN avatar
		_, err = tx.Exec(`UPDATE users SET name = ?, avatar = ? WHERE id = ?`, name, avatarPath, userID)
	} else {
		// Jika tidak ada avatar baru, HANYA update nama
		_, err = tx.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, userID)
	}
	if err != nil {
		return err
	}

	// 2. Update tabel students ATAU lecturers (asumsi user_id unik)
	// Coba update lecturers dulu
	res, err := tx.Exec(`UPDATE lecturers SET address = ?, phone_number = ? WHERE user_id = ?`, address, phone, userID)
	if err != nil { return err }
	
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// Jika tidak ada baris di lecturers yg ter-update, coba update students
		_, err = tx.Exec(`UPDATE students SET address = ?, phone_number = ? WHERE user_id = ?`, address, phone, userID)
		if err != nil { return err }
	}

	return tx.Commit()
}
