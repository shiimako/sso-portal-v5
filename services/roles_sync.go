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

// SyncRoles menarik data dari API Roles
func SyncRoles(env *config.Env, reportFunc func(progress int, msg string), since string) error {
	
	baseURL := env.DataCenterURL
	apiKey  := env.DataCenterKey
	db      := env.DB

	client := &http.Client{Timeout: 10 * time.Second}

	apiURL := fmt.Sprintf("%s/roles/sync", baseURL)
	if since != "" {
		apiURL += fmt.Sprintf("?since=%s", since)
		reportFunc(10, fmt.Sprintf("Mode Delta: %s", since))
	} else {
		reportFunc(10, "Mode Full Sync...")
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil { return fmt.Errorf("req error: %v", err) }

	req.Header.Set("X-API-KEY", apiKey) 

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal koneksi API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API Error: Status Code %d", resp.StatusCode)
	}

	reportFunc(30, "Membaca data role...")

	var result models.DCRoleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("gagal decode JSON: %v", err)
	}

	totalData := len(result.Data)
	if totalData == 0 {
		reportFunc(100, "Data Role kosong.")
		return nil
	}

	reportFunc(50, fmt.Sprintf("Ditemukan %d role. Menyimpan...", totalData))

	savedCount := 0
	failedCount := 0
	var errorDetails []string 

	for i, item := range result.Data {
		currentPercent := 50 + int(float64(i+1)/float64(totalData)*40)

		if item.DeletedAt != nil {
			parsedTime, errParse := time.Parse(time.RFC3339, *item.DeletedAt)
			if errParse == nil {
				mysqlTime := parsedTime.Format("2006-01-02 15:04:05")
				item.DeletedAt = &mysqlTime
			}
		}

		err := models.UpsertRole(db, item)
		
		if err != nil {
			failedCount++
			errDetail := fmt.Sprintf("[ID %d: %v]", item.ID, err)
			errorDetails = append(errorDetails, errDetail)
			
			fmt.Printf("[SYNC ERROR] Role ID %d: %v\n", item.ID, err)
			reportFunc(currentPercent, fmt.Sprintf("❌ Gagal ID %d", item.ID))
		} else {
			savedCount++
			reportFunc(currentPercent, fmt.Sprintf("Menyimpan: %s", item.Name))
		}
		time.Sleep(200 * time.Millisecond) 
	}

	if failedCount > 0 {
		joinedErrors := strings.Join(errorDetails, ", ")

		msg := fmt.Sprintf("⚠️ Selesai Parsial. %d Sukses, %d Gagal.", savedCount, failedCount)
		reportFunc(100, msg)
		return fmt.Errorf("%d role gagal. Detail: %s", failedCount, joinedErrors)
	}

	reportFunc(100, fmt.Sprintf("✅ Selesai. %d Role tersinkronisasi.", savedCount))
	return nil
}