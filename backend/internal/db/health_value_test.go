package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getTestData(r *Repo) *RepoHealthMetrics {
	newMetricsId := uuid.New()
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05.999999999")
	parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", formatted)

	newRepoMetrics := RepoHealthMetrics{
		Id:                  newMetricsId,
		RepoId:              r.Id,
		CreatedAt:           parsedTime,
		ContributerCount:    rand.Intn(100) + 1,
		CommitCount:         rand.Intn(100) + 1,
		SponsorCount:        rand.Intn(100) + 1,
		RepoStarCount:       rand.Intn(100) + 1,
		RepoMultiplierCount: rand.Intn(100) + 1,
		RepoWeight:          rand.Float64(),
	}

	return &newRepoMetrics

}

// done
func TestInsertTrustValue(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	err := InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Nil(t, err)
}

func TestInsertTrustValueDuplicateId(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	err := InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Nil(t, err)
	err = InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Error(t, err)
}

func TestFindRepoHealthMetricsByRepoId(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)

	result, err := FindRepoHealthMetricsByRepoId(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, result[0], *newRepoMetrics)
	assert.Len(t, result, 1)

	newRepoMetrics2 := getTestData(r)
	assert.NotEqual(t, newRepoMetrics, newRepoMetrics2)

	err = InsertRepoHealthMetrics(*newRepoMetrics2)
	assert.Nil(t, err)

	result, err = FindRepoHealthMetricsByRepoId(r.Id)
	assert.Nil(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, result[0].RepoId, result[1].RepoId)
	assert.NotEmpty(t, result[0], result[1])
}

// done
func TestUpdateTrustValue(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	err := InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Nil(t, err)

	alteredRepoMetrics := *newRepoMetrics
	alteredRepoMetrics.RepoMultiplierCount = rand.Intn(100) + 1
	alteredRepoMetrics.ContributerCount = rand.Intn(100) + 1
	alteredRepoMetrics.RepoStarCount = rand.Intn(100) + 1
	assert.NotEqual(t, newRepoMetrics, alteredRepoMetrics)

	err = UpdateRepoHealthMetrics(alteredRepoMetrics)
	assert.Nil(t, err)

	result, err := FindRepoHealthMetricsById(alteredRepoMetrics.Id)
	assert.Nil(t, err)
	assert.Equal(t, alteredRepoMetrics, *result)

}

// done
func TestFindTrustValueById(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)
	result, err := FindRepoHealthMetricsById(newRepoMetrics.Id)
	assert.Nil(t, err)
	assert.Equal(t, newRepoMetrics, result)

}

// done
func TestGetAllTrustValues(t *testing.T) {
	SetupTestData()
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)
	result, err := GetAllTrustValues()
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	for _ = range 5 {
		_ = InsertRepoHealthMetrics(*getTestData(insertTestRepo(t)))
	}
	result, err = GetAllTrustValues()
	assert.Nil(t, err)
	assert.Len(t, result, 6)
}

//func UpdateTrustValue(trustValueMetric RepoHealthMetrics) error {
//	stmt, err := DB.Prepare("UPDATE trust_value_metrics SET repo_id=$1, contributer_count=$2, commit_count=$3, sponsor_donation=$4, sponsor_star_multiplier=$5, repo_sponsor_donated=$6) WHERE id=$7")
//	if err != nil {
//		return fmt.Errorf("prepare UPDATE trust_value_metrics for %v statement failed: %v", trustValueMetric, err)
//	}
//	defer CloseAndLog(stmt)
//
//	var res sql.Result
//	res, err = stmt.Exec(trustValueMetric.Id, trustValueMetric.RepoId, trustValueMetric.ContributerCount, trustValueMetric.CommitCount, trustValueMetric.SponsorCount, trustValueMetric.RepoStarCount, trustValueMetric.RepoMultiplierCount, trustValueMetric.RepoWeight)
//	if err != nil {
//		return err
//	}
//
//	return handleErrMustInsertOne(res)
//}
//
//func FindTrustValueById(id uuid.UUID) (*RepoHealthMetrics, error) {
//	var trustValue RepoHealthMetrics
//	err := DB.
//		QueryRow("SELECT id, repo_id, created_at, contributer_count, commit_count,  sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated from trust_value WHERE id=$1", id).
//		Scan(&trustValue.Id, &trustValue.RepoId, &trustValue.CreatedAt, &trustValue.ContributerCount, &trustValue.CommitCount, &trustValue.SponsorCount, &trustValue.RepoStarCount, &trustValue.RepoMultiplierCount)
//	if err != nil {
//		return nil, err
//	}
//
//	switch err {
//	case sql.ErrNoRows:
//		return nil, nil
//	case nil:
//		return &trustValue, nil
//	default:
//		return nil, err
//	}
//}
//
//func FindTrustValueByRepoId(repoId uuid.UUID) ([]RepoHealthMetrics, error) {
//	//var tv TrustValue
//	rows, err := DB.
//		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated from trust_value WHERE repo_id=$1 order by created_at desc limit 1", repoId)
//	if err != nil {
//		return nil, err
//	}
//	defer CloseAndLog(rows)
//	return scanTrustValue(rows)
//}
//
//func scanTrustValue(rows *sql.Rows) ([]RepoHealthMetrics, error) {
//	trustValues := []RepoHealthMetrics{}
//	for rows.Next() {
//		var tv RepoHealthMetrics
//		err := rows.Scan(&tv.Id, &tv.RepoId, &tv.CreatedAt, &tv.ContributerCount, &tv.CommitCount, &tv.SponsorCount, &tv.RepoStarCount, &tv.RepoMultiplierCount, &tv.RepoWeight)
//		if err != nil {
//			return nil, err
//		}
//		trustValues = append(trustValues, tv)
//	}
//	return trustValues, nil
//}
//
//func GetAllTrustValues() ([]RepoHealthMetrics, error) {
//	rows, err := DB.
//		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated from trust_value order by created_at desc")
//	if err != nil {
//		return nil, err
//	}
//
//	return scanTrustValue(rows)
//}
//
