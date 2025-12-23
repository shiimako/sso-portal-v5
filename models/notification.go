package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type RawNotification struct {
	AppID   int    `db:"app_id"`
	Slug    string `db:"slug"`
	AppName string `db:"app_name"`
	Message string `db:"message"`
}

type NotifSummary struct {
	Count    int
	Messages []string
}

// InsertNotification: Selalu INSERT baris baru (Stacking)
func InsertNotification(db *sqlx.DB, userID int, appID int, msg string) error {
	query := `INSERT INTO application_notifications (user_id, app_id, message, updated_at) VALUES (?, ?, ?, NOW())`
	_, err := db.Exec(query, userID, appID, msg)
	return err
}

// GetNotificationSummary: Mengelompokkan pesan berdasarkan Aplikasi
func GetNotificationSummary(db *sqlx.DB, userID int) (map[string]NotifSummary, error) {
	var rows []RawNotification
	
	query := `
		SELECT 
			an.app_id, 
			a.slug, 
			a.name as app_name,
			an.message
		FROM application_notifications an
		JOIN applications a ON an.app_id = a.id
		WHERE an.user_id = ?
		ORDER BY an.updated_at DESC
	`
	
	err := db.Select(&rows, query, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	summary := make(map[string]NotifSummary)

	for _, row := range rows {
		data := summary[row.Slug]
		
		data.Count++
		data.Messages = append(data.Messages, row.Message)
		
		summary[row.Slug] = data
	}

	return summary, nil
}

// ClearNotification: Hapus SEMUA notif untuk app tertentu
func ClearNotification(db *sqlx.DB, userID int, appID int) error {
	query := "DELETE FROM application_notifications WHERE user_id = ? AND app_id = ?"
	_, err := db.Exec(query, userID, appID)
	return err
}