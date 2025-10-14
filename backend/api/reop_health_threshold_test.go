package api

import (
	"backend/db"
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

	jsonData, err := json.Marshal(testData)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/thresholds", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	SetNewThresholds(w, req, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var response db.RepoHealthThreshold
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.Id)
	assert.NotEmpty(t, response.CreatedAt)
	assert.Equal(t, 0, response.ThContributorCount.Lower)
	assert.Equal(t, 100, response.ThContributorCount.Upper)
	assert.Equal(t, 0, response.ThCommitCount.Lower)
	assert.Equal(t, 500, response.ThCommitCount.Upper)
}

func TestValidateThresholds(t *testing.T) {
	createValidThresholds := func() *db.RepoHealthThreshold {
		return &db.RepoHealthThreshold{
			ThContributorCount:   &db.Threshold{Upper: 100, Lower: 0},
			ThCommitCount:        &db.Threshold{Upper: 500, Lower: 0},
			ThSponsorDonation:    &db.Threshold{Upper: 1000, Lower: 0},
			ThRepoStarCount:      &db.Threshold{Upper: 1000, Lower: 0},
			ThRepoMultiplier:     &db.Threshold{Upper: 10, Lower: 0},
			ThActiveFFSUserCount: &db.Threshold{Upper: 50, Lower: 0},
		}
	}

	t.Run("Valid thresholds", func(t *testing.T) {
		tests := []struct {
			name      string
			threshold *db.RepoHealthThreshold
		}{
			{
				name: "all zeros",
				threshold: &db.RepoHealthThreshold{
					ThContributorCount:   &db.Threshold{Upper: 0, Lower: 0},
					ThCommitCount:        &db.Threshold{Upper: 0, Lower: 0},
					ThSponsorDonation:    &db.Threshold{Upper: 0, Lower: 0},
					ThRepoStarCount:      &db.Threshold{Upper: 0, Lower: 0},
					ThRepoMultiplier:     &db.Threshold{Upper: 0, Lower: 0},
					ThActiveFFSUserCount: &db.Threshold{Upper: 0, Lower: 0},
				},
			},
			{
				name: "equal upper and lower",
				threshold: &db.RepoHealthThreshold{
					ThContributorCount:   &db.Threshold{Upper: 5, Lower: 5},
					ThCommitCount:        &db.Threshold{Upper: 5, Lower: 5},
					ThSponsorDonation:    &db.Threshold{Upper: 5, Lower: 5},
					ThRepoStarCount:      &db.Threshold{Upper: 5, Lower: 5},
					ThRepoMultiplier:     &db.Threshold{Upper: 5, Lower: 5},
					ThActiveFFSUserCount: &db.Threshold{Upper: 5, Lower: 5},
				},
			},
			{
				name:      "valid range",
				threshold: createValidThresholds(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateThresholds(tt.threshold)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("Missing fields", func(t *testing.T) {
		tests := []struct {
			name      string
			threshold *db.RepoHealthThreshold
		}{
			{
				name: "nil contributor count",
				threshold: func() *db.RepoHealthThreshold {
					th := createValidThresholds()
					th.ThContributorCount = nil
					return th
				}(),
			},
			{
				name: "nil commit count",
				threshold: func() *db.RepoHealthThreshold {
					th := createValidThresholds()
					th.ThCommitCount = nil
					return th
				}(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateThresholds(tt.threshold)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "all threshold fields are required")
			})
		}
	})

	t.Run("Negative values", func(t *testing.T) {
		tests := []struct {
			name      string
			threshold *db.RepoHealthThreshold
			field     string
		}{
			{
				name: "negative lower contributor count",
				threshold: func() *db.RepoHealthThreshold {
					th := createValidThresholds()
					th.ThContributorCount.Lower = -1
					return th
				}(),
				field: "contributor_count",
			},
			{
				name: "negative upper contributor count",
				threshold: func() *db.RepoHealthThreshold {
					th := createValidThresholds()
					th.ThContributorCount.Upper = -1
					return th
				}(),
				field: "contributor_count",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateThresholds(tt.threshold)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot be negative")
				assert.Contains(t, err.Error(), tt.field)
			})
		}
	})

	t.Run("Upper less than lower", func(t *testing.T) {
		tests := []struct {
			name      string
			threshold *db.RepoHealthThreshold
			field     string
		}{
			{
				name: "contributor count upper < lower",
				threshold: func() *db.RepoHealthThreshold {
					th := createValidThresholds()
					th.ThContributorCount.Upper = 5
					th.ThContributorCount.Lower = 10
					return th
				}(),
				field: "contributor_count",
			},
			{
				name: "commit count upper < lower",
				threshold: func() *db.RepoHealthThreshold {
					th := createValidThresholds()
					th.ThCommitCount.Upper = 5
					th.ThCommitCount.Lower = 10
					return th
				}(),
				field: "commit_count",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateThresholds(tt.threshold)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "upper threshold cannot be less than lower threshold")
				assert.Contains(t, err.Error(), tt.field)
			})
		}
	})
}
