package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sso-portal-v3/models"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// SyncUsers menarik data User (Support Pagination)
func SyncUsers(db *sqlx.DB, reportFunc func(progress int, msg string)) error {
	
	baseURL := "http://127.0.0.1:9999/api/v1/users/sync"
	client := &http.Client{Timeout: 30 * time.Second}
	limit := 50 
	offset := 0 
	page := 1
	
	savedCount := 0
	failedCount := 0
	var errorDetails []string

	reportFunc(5, "Memulai sinkronisasi User...")

	for {
		apiURL := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, offset)
		
		reportFunc(10+(page), fmt.Sprintf("Fetching Page %d...", page))

		resp, err := client.Get(apiURL)
		if err != nil {
			return fmt.Errorf("koneksi gagal: %v", err)
		}
		
		var result models.DCUserResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return fmt.Errorf("json decode error: %v", err)
		}
		resp.Body.Close()

		if len(result.Data) == 0 {
			break
		}

		for _, item := range result.Data {
			err := models.UpsertFullUser(db, item)
			if err != nil {
				failedCount++
				errDetail := fmt.Sprintf("[User: %s, Error: %v]", item.Email, err)
				errorDetails = append(errorDetails, errDetail)
				fmt.Println("Sync Error:", errDetail)
			} else {
				savedCount++
			}
		}

		reportFunc(10+(page*2), fmt.Sprintf("Page %d selesai (%d Tersimpan, %d Gagal)", page, savedCount, failedCount))

		if len(result.Data) < limit {
			break 
		}
		offset += limit
		page++
		time.Sleep(200 * time.Millisecond)
	}

	if failedCount > 0 {
		joinedErrors := strings.Join(errorDetails, "; ")
		
		reportFunc(100, fmt.Sprintf("⚠️ Selesai Parsial. %d Sukses, %d Gagal.", savedCount, failedCount))
		return fmt.Errorf("%d user gagal. Detail: %s", failedCount, joinedErrors)
	}

	reportFunc(100, fmt.Sprintf("✅ Selesai. Total %d User tersinkronisasi.", savedCount))
	return nil
}