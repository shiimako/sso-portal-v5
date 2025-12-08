package models

import "github.com/jmoiron/sqlx"

// GetLastSuccessTime mengambil waktu created_at dari log sukses terakhir module USER
func GetLastSuccessTime(db *sqlx.DB, module string) (string, error) {
	var lastTime string
	// Ambil log sukses terakhir, format ke ISO 8601 (RFC3339)
	query := `SELECT DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ') FROM sync_logs WHERE module = ? AND status = 'SUCCESS' ORDER BY created_at DESC LIMIT 1`
	err := db.Get(&lastTime, query, module)
	return lastTime, err
}
