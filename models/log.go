package models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type SyncLog struct {
	ID        int       `db:"id"`
	Type      string    `db:"type"`      // MANUAL, CRON, WEBHOOK
	Module    string    `db:"module"`    
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

// GetLogs dengan Filter & Sorting DESC
func GetLogs(db *sqlx.DB, limit, offset int, moduleFilter string) ([]SyncLog, error) {
	var logs []SyncLog
	var err error

	query := "SELECT * FROM sync_logs"
	args := []interface{}{}

	if moduleFilter != "" {
		query += " WHERE module = ?"
		args = append(args, moduleFilter)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err = db.Select(&logs, query, args...)
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

// CountLogs dengan Filter (untuk Pagination)
func CountLogs(db *sqlx.DB, moduleFilter string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM sync_logs"
	args := []interface{}{}

	if moduleFilter != "" {
		query += " WHERE module = ?"
		args = append(args, moduleFilter)
	}

	err := db.Get(&count, query, args...)
	return count, err
}