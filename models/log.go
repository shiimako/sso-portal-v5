package models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type SyncLog struct {
	ID        int       `db:"id"`
	Type      string    `db:"type"`      // MANUAL, CRON, WEBHOOK
	Module    string    `db:"module"`    // USER
	Status    string    `db:"status"`    // RUNNING, SUCCESS, ERROR
	Message   string    `db:"message"`
	IsRead    bool      `db:"is_read"`
	CreatedAt time.Time `db:"created_at"`
}

// CreateLog mencatat log baru
func CreateLog(db *sqlx.DB, logType, module, status, message string) error {
	query := `INSERT INTO sync_logs (type, module, status, message, is_read) VALUES (?, ?, ?, ?, ?)`
	// Jika status ERROR, is_read = false (biar muncul badge), selain itu anggap read
	isRead := status != "ERROR" 
	_, err := db.Exec(query, logType, module, status, message, isRead)
	return err
}

// GetLogs mengambil list log (pagination)
func GetLogs(db *sqlx.DB, limit, offset int) ([]SyncLog, error) {
	var logs []SyncLog
	query := `SELECT * FROM sync_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := db.Select(&logs, query, limit, offset)
	return logs, err
}

// CountUnreadErrors menghitung error yang belum dibaca (untuk Badge Header/Dashboard)
func CountUnreadErrors(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM sync_logs WHERE status = 'ERROR' AND is_read = 0")
	return count, err
}

// MarkErrorsAsRead menandai semua error sebagai sudah dibaca
func MarkErrorsAsRead(db *sqlx.DB) error {
	_, err := db.Exec("UPDATE sync_logs SET is_read = 1 WHERE status = 'ERROR'")
	return err
}