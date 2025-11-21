// file: models/application.go

package models

import (
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Application mendefinisikan struktur data untuk sebuah aplikasi klien.
type Application struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	Slug      string `db:"slug"`
	TargetURL string `db:"target_url"`
}

var cacheRoleApps = map[string][]Application{}
var cachePosApps = map[string][]Application{}

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

func FindApplicationsByRole(db *sqlx.DB, roleName string) ([]Application, error) {

	if apps, ok := cacheRoleApps[roleName]; ok {
		return apps, nil
	}

	var apps []Application
	query := `
	SELECT a.id, a.name, a.slug, a.target_url
	FROM applications a
	JOIN application_role_access aa ON a.id = aa.application_id
	JOIN roles r ON aa.role_id = r.id
	WHERE r.name = ?`
	err := db.Select(&apps, query, roleName)
	return apps, err
}

func FindApplicationsByPositions(db *sqlx.DB, positionIDs []int) ([]Application, error) {

	if len(positionIDs) == 0 {
		return []Application{}, nil
	}

	key := sliceKey(positionIDs)

	if apps, ok := cachePosApps[key]; ok {
		return apps, nil
	}

	var apps []Application
	query := `
	SELECT DISTINCT a.id, a.name, a.slug, a.target_url
	FROM applications a
	JOIN application_position_access aa ON a.id = aa.application_id
	WHERE aa.position_id IN (?)`
	query, args, err := sqlx.In(query, positionIDs)
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)
	err = db.Select(&apps, query, args...)
	if err != nil {
		return nil, err
	}
	cachePosApps[key] = apps

	return apps, nil
}

func sliceKey(ints []int) string {
    key := make([]string, len(ints))
    for i, v := range ints {
        key[i] = strconv.Itoa(v)
    }
    return strings.Join(key, ",")
}

