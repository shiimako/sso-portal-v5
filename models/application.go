// file: models/application.go

package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Application mendefinisikan struktur data untuk sebuah aplikasi klien.
type Application struct {
	ID          int            `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	Slug        string         `db:"slug"`
	TargetURL   string         `db:"target_url"`
	IconURL     sql.NullString `db:"icon_url"`
}

// GetAllApplications mengambil semua data aplikasi dari database.
func GetAllApplications(db *sqlx.DB, page int, pagesize int, search string) ([]Application, error) {
	offset := (page - 1) * pagesize

	apps := []Application{}

	query := `
		SELECT id, name, description, slug, target_url, icon_url
		FROM applications
		WHERE 1=1
	`

	params := []interface{}{}

	if search != "" {
		query += ` AND (name LIKE ? OR slug LIKE ? OR description LIKE ?) `
		like := "%" + search + "%"
		params = append(params, like, like, like)
	}

	// Ordering & pagination
	query += ` ORDER BY id ASC LIMIT ? OFFSET ? `
	params = append(params, pagesize, offset)

	// Execute query
	err := db.Select(&apps, query, params...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

// CreateApplication menyimpan aplikasi baru dan hak akses perannya dalam satu transaksi.
func CreateApplication(db *sqlx.DB, name, description, slug, targetURL, iconURL string, roleIDs []string, positionIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`INSERT INTO applications (name, description, slug, target_url, icon_url) VALUES (?, ?, ?, ?, ?)`,
		name, description, slug, targetURL, iconURL)
	if err != nil {
		return err
	}
	appID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if len(roleIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO application_role_access (application_id, role_id) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, rid := range roleIDs {
			_, err := stmt.Exec(appID, rid)
			if err != nil {
				return err
			}
		}
	}

	if len(positionIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO application_position_access (application_id, position_id) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, pid := range positionIDs {
			_, err := stmt.Exec(appID, pid)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// FindApplicationByID mengambil satu aplikasi dan daftar ID peran yang terkait.
func FindApplicationByID(db *sqlx.DB, id string) (Application, []int, []int, error) {
	var app Application
	var roleIDs []int
	var posIDs []int

	// 1. Ambil detail aplikasi
	queryApp := `SELECT id, name, description, slug, target_url, icon_url FROM applications WHERE id = ?`
	err := db.QueryRow(queryApp, id).Scan(&app.ID, &app.Name, &app.Description, &app.Slug, &app.TargetURL, &app.IconURL)
	if err != nil {
		return app, nil, nil, err
	}

	// 2. Ambil semua role_id yang terhubung dengan aplikasi ini
	queryRoles := `SELECT role_id FROM application_role_access WHERE application_id = ?`
	rows1, err := db.Query(queryRoles, id)
	if err != nil {
		return app, nil, nil, err
	}
	defer rows1.Close()

	for rows1.Next() {
		var roleID int
		if err := rows1.Scan(&roleID); err != nil {
			continue
		}
		roleIDs = append(roleIDs, roleID)
	}

	queryPosition := `SELECT position_id FROM application_position_access WHERE application_id = ?`
	rows2, err := db.Query(queryPosition, id)
	if err != nil {
		return app, nil, nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var posID int
		if err := rows2.Scan(&posID); err != nil {
			continue
		}
		posIDs = append(posIDs, posID)
	}

	return app, roleIDs, posIDs, nil
}

// UpdateApplication memperbarui data aplikasi dan hak akses perannya dalam satu transaksi.
func UpdateApplication(db *sqlx.DB, id, name, description, slug, targetURL, iconURL string, roleIDs []string, posIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE applications SET name=?, description =?, slug=?, target_url=?, icon_url =? WHERE id=?`,
		name, description, slug, targetURL, iconURL, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM application_role_access WHERE application_id=?`, id)
	if err != nil {
		return err
	}

	if len(roleIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO application_role_access (application_id, role_id) VALUES (?, ?)`)
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

	_, err = tx.Exec(`DELETE FROM application_position_access WHERE application_id=?`, id)
	if err != nil {
		return err
	}

	if len(posIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO application_position_access (application_id, position_id) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, pid := range posIDs {
			_, err := stmt.Exec(id, pid)
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeleteApplication menghapus aplikasi dan semua hak akses terkait dari database.
func DeleteApplication(db *sqlx.DB, id string) error {
	_, err := db.Exec(`DELETE FROM applications WHERE id=?`, id)
	if err != nil {
		return err
	}

	return nil
}

// FindApplicationBySlug mengambil satu aplikasi berdasarkan slug-nya.
func FindApplicationBySlug(db *sqlx.DB, slug string) (Application, error) {
	var app Application
	query := `SELECT id, name, description, slug, target_url, icon_url FROM applications WHERE slug = ?`
	err := db.Get(&app, query, slug)
	return app, err
}

// FindAccessibleApps mengambil aplikasi yang dapat diakses berdasarkan role dan posisi.
func FindAccessibleApps(db *sqlx.DB, roleName string, positionIDs []int) ([]Application, error) {
    
    // TRICK: Handle Empty Slice
    // sqlx.In akan error jika slice kosong. 
    // Jadi jika kosong, kita isi string kosong "" agar query menjadi: IN ('')
    // Ini aman karena tidak ada position_name yang namanya kosong.
    if len(positionIDs) == 0 {
        positionIDs = []int{0}
    }

    query := `
    -- Bagian 1: Ambil Apps berdasarkan ROLE (Admin/Mhs/Dosen)
    SELECT a.id, a.name, a.description, a.slug, a.target_url, a.icon_url
    FROM applications a
    JOIN application_role_access ara ON a.id = ara.application_id
    JOIN roles r ON ara.role_id = r.id
    WHERE r.role_name = ?
    
    UNION

    -- Bagian 2: Ambil Apps berdasarkan POSITION (untuk Dosen)
    SELECT a.id, a.name, a.description, a.slug, a.target_url, a.icon_url
    FROM applications a
    JOIN application_position_access apa ON a.id = apa.application_id
    WHERE apa.position_id IN (?)
    `

    // Proses Query
    query, args, err := sqlx.In(query, roleName, positionIDs)
    if err != nil {
        return nil, err
    }
    
    query = db.Rebind(query)
    
    var apps []Application
    err = db.Select(&apps, query, args...)
    
    return apps, err
}
