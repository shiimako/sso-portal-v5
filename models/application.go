// file: models/application.go

package models

import (
	"github.com/jmoiron/sqlx"
)

// Application mendefinisikan struktur data untuk sebuah aplikasi klien.
type Application struct {
	ID        int
	Name      string
	Slug      string
	TargetURL string `db:"target_url"`
}

// GetAllApplications mengambil semua data aplikasi dari database.
func GetAllApplications(db *sqlx.DB) ([]Application, error) {
	var apps []Application

	query := `SELECT id, name, slug, target_url FROM applications ORDER BY id ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var app Application
		if err := rows.Scan(&app.ID, &app.Name, &app.Slug, &app.TargetURL); err != nil {
			continue
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// CreateApplication menyimpan aplikasi baru dan hak akses perannya dalam satu transaksi.
func CreateApplication(db *sqlx.DB, name, slug, targetURL string, roleIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback jika ada panic atau error yang tidak ditangani

	// 1. Insert ke tabel applications
	result, err := tx.Exec(`INSERT INTO applications (name, slug, target_url) VALUES (?, ?, ?)`,
		name, slug, targetURL)
	if err != nil {
		return err
	}

	// Dapatkan ID aplikasi yang baru saja dibuat
	appID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// 2. Insert hak akses ke tabel application_access jika ada peran yang dipilih
	if len(roleIDs) > 0 {
		// Siapkan statement untuk efisiensi jika banyak peran
		stmt, err := tx.Prepare(`INSERT INTO application_access (application_id, role_id) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, rid := range roleIDs {
			_, err := stmt.Exec(appID, rid)
			if err != nil {
				return err // Jika salah satu gagal, seluruh transaksi akan di-rollback
			}
		}
	}

	// Jika semua berhasil, commit transaksinya
	return tx.Commit()
}

// FindApplicationByID mengambil satu aplikasi dan daftar ID peran yang terkait.
func FindApplicationByID(db *sqlx.DB, id string) (Application, []int, error) {
	var app Application
	var roleIDs []int

	// 1. Ambil detail aplikasi
	queryApp := `SELECT id, name, slug, target_url FROM applications WHERE id = ?`
	err := db.QueryRow(queryApp, id).Scan(&app.ID, &app.Name, &app.Slug, &app.TargetURL)
	if err != nil {
		return app, nil, err
	}

	// 2. Ambil semua role_id yang terhubung dengan aplikasi ini
	queryRoles := `SELECT role_id FROM application_access WHERE application_id = ?`
	rows, err := db.Query(queryRoles, id)
	if err != nil {
		return app, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleID int
		if err := rows.Scan(&roleID); err != nil {
			continue
		}
		roleIDs = append(roleIDs, roleID)
	}

	return app, roleIDs, nil
}

// UpdateApplication memperbarui data aplikasi dan hak akses perannya dalam satu transaksi.
func UpdateApplication(db *sqlx.DB, id, name, slug, targetURL string, roleIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update tabel applications
	_, err = tx.Exec(`UPDATE applications SET name=?, slug=?, target_url=? WHERE id=?`,
		name, slug, targetURL, id)
	if err != nil {
		return err
	}

	// 2. Hapus semua hak akses lama untuk aplikasi ini
	_, err = tx.Exec(`DELETE FROM application_access WHERE application_id=?`, id)
	if err != nil {
		return err
	}

	// 3. Masukkan hak akses baru
	if len(roleIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO application_access (application_id, role_id) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, rid := range roleIDs {
			_, err := stmt.Exec(id, rid)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// DeleteApplication menghapus aplikasi dan semua hak akses terkait dari database.
func DeleteApplication(db *sqlx.DB, id string) error {
	_, err := db.Exec(`DELETE FROM applications WHERE id=?`, id)
	return err
}

// FindApplicationBySlug mengambil satu aplikasi berdasarkan slug-nya.
func FindApplicationBySlug(db *sqlx.DB, slug string) (Application, error) {
	var app Application
	query := `SELECT id, name, slug, target_url FROM applications WHERE slug = ?`
	err := db.Get(&app, query, slug)
	return app, err
}
