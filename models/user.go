package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// UserWithRoles adalah struct untuk menampilkan daftar pengguna di tabel admin.
type UserWithRoles struct {
	ID     int
	Name   string
	Email  string
	Status string
	Roles  string
}

// User adalah struct untuk data user lengkap.
type User struct {
	ID         int
	Name       string
	Email      string
	Status     string
	Roles      []int          // Slice of role IDs
	RolesStr   string         // Comma-separated role IDs for easier checking in templates
	Attributes map[int]string // Map of attribute role IDs for checkbox checking

	NIM  sql.NullString
	NIP  sql.NullString
	NIDN sql.NullString

	Address sql.NullString
	Phone   sql.NullString
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

	rows, err := db.Query(query)
	if err != nil {
		return nil, err // Kembalikan error jika query gagal
	}
	defer rows.Close()

	for rows.Next() {
		var user UserWithRoles
		var roles sql.NullString
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Status, &roles); err != nil {
			// Jika ada error saat scan, lanjutkan ke baris berikutnya
			continue
		}
		user.Roles = roles.String
		users = append(users, user)
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
	err := db.Get(&user, `SELECT id, name, email, status FROM users WHERE id=?`, id)
	if err != nil {
		return user, err
	}

	type UserRoleLink struct {
		RoleID   int            `db:"id"`
		RoleType string         `db:"type"`
		Scope    sql.NullString `db:"scope"` // Gunakan NullString untuk scan
	}
	var assignedRoles []UserRoleLink
	err = db.Select(&assignedRoles, `SELECT r.id, r.type, ur.scope FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = ?`, id)
	if err != nil && err != sql.ErrNoRows {
		return user, err
	}

	var hasStudentRole, hasLecturerRole bool
	user.Attributes = make(map[int]string)
	var sb strings.Builder
	sb.WriteString(",")

	for _, r := range assignedRoles {
		if r.RoleType == "base" {
			user.Roles = append(user.Roles, r.RoleID)
			sb.WriteString(fmt.Sprintf("%d,", r.RoleID))
			if r.RoleID == 2 {
				hasLecturerRole = true
			}
			if r.RoleID == 3 {
				hasStudentRole = true
			}
		} else if r.RoleType == "attribute" {
			user.Attributes[r.RoleID] = r.Scope.String // [FIX] Gunakan .String
		}
	}
	user.RolesStr = sb.String()

	if hasStudentRole {
		_ = db.QueryRow(`SELECT nim, address, phone_number FROM students WHERE user_id=?`, id).Scan(&user.NIM, &user.Address, &user.Phone)
	} else if hasLecturerRole {
		_ = db.QueryRow(`SELECT nip, nidn, address, phone_number FROM lecturers WHERE user_id=?`, id).Scan(&user.NIP, &user.NIDN, &user.Address, &user.Phone)
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
