package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sso-portal-v3/config"
	"sso-portal-v3/models"
	"strings"
	"time"
)

// SyncUsers menarik data User (Support Pagination)
func SyncUsers(env *config.Env, reportFunc func(progress int, msg string), since string) error {

	baseURL := fmt.Sprintf("%s/users/sync", env.DataCenterURL)
	apiKey := env.DataCenterKey
	db := env.DB

	client := &http.Client{Timeout: 10 * time.Second}
	limit := 50
	offset := 0
	page := 1

	savedCount := 0
	failedCount := 0
	var errorDetails []string

	reportFunc(5, "Memulai sinkronisasi User...")

	for {
		apiURL := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, offset)
		if since != "" {
			apiURL += fmt.Sprintf("&since=%s", since)
			reportFunc(10, fmt.Sprintf("Mode Delta: %s", since))
		} else {
			reportFunc(10, "Mode Full Sync...")
		}

		reportFunc(10+(page), fmt.Sprintf("Fetching Page %d...", page))

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return fmt.Errorf("req error: %v", err)
		}

		req.Header.Set("X-API-KEY", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("gagal koneksi API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("API Error: Status Code %d", resp.StatusCode)
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
