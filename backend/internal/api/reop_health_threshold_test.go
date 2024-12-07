package api

import (
	"backend/internal/db"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNewThresholds_Success(t *testing.T) {
	db.SetupTestData()
	defer db.TeardownTestData()
	// Create test data using the actual structs
	testData := db.RepoHealthThreshold{
		ThContributorCount: &db.Threshold{
			Upper: 100,
			Lower: 0,
		},
		ThCommitCount: &db.Threshold{
			Upper: 500,
			Lower: 0,
		},
		ThSponsorDonation: &db.Threshold{
			Upper: 1000,
			Lower: 0,
		},
		ThRepoStarCount: &db.Threshold{
			Upper: 1000,
			Lower: 0,
		},
		ThRepoMultiplier: &db.Threshold{
			Upper: 10,
			Lower: 0,
		},
		ThActiveFFSUserCount: &db.Threshold{
			Upper: 50,
			Lower: 0,
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(testData)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/thresholds", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	SetNewThresholds(w, req, nil)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response db.RepoHealthThreshold
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Verify response fields
	assert.NotEmpty(t, response.Id)
	assert.NotEmpty(t, response.CreatedAt)
	assert.Equal(t, 0, response.ThContributorCount.Lower)
	assert.Equal(t, 100, response.ThContributorCount.Upper)
	assert.Equal(t, 0, response.ThCommitCount.Lower)
	assert.Equal(t, 500, response.ThCommitCount.Upper)
}
