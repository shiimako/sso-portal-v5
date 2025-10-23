package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func InitDB() (*sqlx.DB, error) {
	// Ambil konfigurasi dari environment
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	// Bikin connection string MySQL (user:pass@tcp(host)/dbname)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbHost, dbName)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	// Lakukan ping untuk memastikan database benar-benar terhubung
	err = db.Ping()
	if err != nil {
		db.Close() // Tutup koneksi jika ping gagal
		return nil, fmt.Errorf("database tidak dapat dijangkau: %w", err)
	}

	log.Println("Koneksi ke database ", dbName, " berhasil.")
	return db, nil
}