package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sso-portal-v3/models"
	"time"
	"strings"

	"github.com/jmoiron/sqlx"
)

// SyncMajors menarik data dari API Jurusan
func SyncMajors(db *sqlx.DB, reportFunc func(progress int, msg string)) error {

	baseURL := "http://127.0.0.1:9999/api/v1/jurusan/sync" // Sesuaikan port
	client := &http.Client{Timeout: 10 * time.Second}

	reportFunc(10, "Menghubungkan ke API Data Center...")

	resp, err := client.Get(baseURL)
	if err != nil {
		return fmt.Errorf("gagal koneksi API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API Error: Status Code %d", resp.StatusCode)
	}

	reportFunc(30, "Membaca data...")

	var result models.DCMajorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("gagal decode JSON: %v", err)
	}

	totalData := len(result.Data)
	if totalData == 0 {
		reportFunc(100, "Data Jurusan kosong dari API.")
		return nil
	}

	reportFunc(50, fmt.Sprintf("Ditemukan %d jurusan. Menyimpan...", totalData))

	savedCount := 0
	failedCount := 0
	var errorDetails []string // 1. Penampung detail error

	for i, item := range result.Data {
		currentPercent := 50 + int(float64(i+1)/float64(totalData)*40)

		// Parsing Datetime
		if item.DeletedAt != nil {
			parsedTime, errParse := time.Parse(time.RFC3339, *item.DeletedAt)
			if errParse == nil {
				mysqlTime := parsedTime.Format("2006-01-02 15:04:05")
				item.DeletedAt = &mysqlTime
			}
		}

		err := models.UpsertMajor(db, item)

		if err != nil {
			failedCount++

			// 2. Simpan detail error
			errDetail := fmt.Sprintf("[ID %d: %v]", item.ID, err)
			errorDetails = append(errorDetails, errDetail)

			// Log console
			fmt.Printf("[SYNC ERROR] Jurusan ID %d: %v\n", item.ID, err)
			reportFunc(currentPercent, fmt.Sprintf("❌ Gagal ID %d", item.ID))
		} else {
			savedCount++
			reportFunc(currentPercent, fmt.Sprintf("Menyimpan: %s", item.Name))
		}

		time.Sleep(100 * time.Millisecond)
	}

	// 3. Logic Return Error Lengkap
	if failedCount > 0 {
		// Gabungkan error message (batasi panjang jika perlu)
		joinedErrors := strings.Join(errorDetails, ", ")

		msg := fmt.Sprintf("⚠️ Selesai Parsial. %d Sukses, %d Gagal.", savedCount, failedCount)
		reportFunc(100, msg)

		// Return error detail ke Controller -> Database Log
		return fmt.Errorf("%d jurusan gagal. Detail: %s", failedCount, joinedErrors)
	}

	reportFunc(100, fmt.Sprintf("✅ Selesai. %d Jurusan tersinkronisasi.", savedCount))
	return nil
}
