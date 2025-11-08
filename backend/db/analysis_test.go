package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertAnalysisRequest(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/analysis-repo")

	analysisRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/analysis-repo",
	}

	err := db.InsertAnalysisRequest(analysisRequest, time.Now())
	require.NoError(t, err)
}

func TestInsertRepoMetric(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/analysis-repo")

	analysisRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/analysis-repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(analysisRequest, time.Now()))

	gitEmail := "contributor@example.com"
	names := []string{"Contributor Name", "Another Name"}
	weight := 0.75

	err := db.InsertRepoMetric(analysisRequest.Id, repo.Id, gitEmail, names, weight, time.Now())
	require.NoError(t, err)
}

func TestFindLatestAnalysisRequest(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/analysis-repo")

	oldRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -60),
		DateTo:   time.Now().AddDate(0, 0, -30),
		GitUrl:   "https://github.com/test/analysis-repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(oldRequest, time.Now()))

	latestRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/analysis-repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(latestRequest, time.Now()))

	found, err := db.FindLatestAnalysisRequest(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, latestRequest.Id, found.Id)
}

func TestFindLatestAnalysisRequest_NotFound(t *testing.T) {
	TruncateAll(db, t)

	found, err := db.FindLatestAnalysisRequest(uuid.New())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindAllLatestAnalysisRequest(t *testing.T) {
	TruncateAll(db, t)

	repo1 := createTestRepo(t, db, "https://github.com/test/repo1")
	repo2 := createTestRepo(t, db, "https://github.com/test/repo2")

	dateTo := time.Now()

	request1 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo1.Id,
		DateFrom: dateTo.AddDate(0, 0, -30),
		DateTo:   dateTo,
		GitUrl:   "https://github.com/test/repo1",
	}
	require.NoError(t, db.InsertAnalysisRequest(request1, time.Now()))

	request2 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo2.Id,
		DateFrom: dateTo.AddDate(0, 0, -30),
		DateTo:   dateTo,
		GitUrl:   "https://github.com/test/repo2",
	}
	require.NoError(t, db.InsertAnalysisRequest(request2, time.Now()))

	found, err := db.FindAllLatestAnalysisRequest(dateTo.Add(time.Hour))
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestFindAllLatestAnalysisRequest_FiltersByDateTo(t *testing.T) {
	TruncateAll(db, t)

	repo1 := createTestRepo(t, db, "https://github.com/test/repo1")
	repo2 := createTestRepo(t, db, "https://github.com/test/repo2")

	oldDateTo := time.Now().AddDate(0, 0, -10)
	request1 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo1.Id,
		DateFrom: oldDateTo.AddDate(0, 0, -30),
		DateTo:   oldDateTo,
		GitUrl:   "https://github.com/test/repo1",
	}
	require.NoError(t, db.InsertAnalysisRequest(request1, time.Now()))

	newDateTo := time.Now()
	request2 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo2.Id,
		DateFrom: newDateTo.AddDate(0, 0, -30),
		DateTo:   newDateTo,
		GitUrl:   "https://github.com/test/repo2",
	}
	require.NoError(t, db.InsertAnalysisRequest(request2, time.Now()))

	found, err := db.FindAllLatestAnalysisRequest(oldDateTo.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, request1.Id, found[0].Id)
}

func TestUpdateAnalysisRequest(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	request := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(request, time.Now()))

	receivedAt := time.Now()
	errStr := "test error"
	err := db.UpdateAnalysisRequest(request.Id, receivedAt, &errStr)
	require.NoError(t, err)

	found, err := db.FindLatestAnalysisRequest(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, found.ReceivedAt)
	require.NotNil(t, found.Error)
	assert.Equal(t, errStr, *found.Error)
}

func TestUpdateAnalysisRequest_NoError(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	request := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(request, time.Now()))

	receivedAt := time.Now()
	err := db.UpdateAnalysisRequest(request.Id, receivedAt, nil)
	require.NoError(t, err)

	found, err := db.FindLatestAnalysisRequest(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, found.ReceivedAt)
	assert.Nil(t, found.Error)
}

func TestFindAnalysisResults(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	request := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(request, time.Now()))

	gitEmail1 := "contributor1@example.com"
	names1 := []string{"Contributor One"}
	weight1 := 0.6
	require.NoError(t, db.InsertRepoMetric(request.Id, repo.Id, gitEmail1, names1, weight1, time.Now()))

	gitEmail2 := "contributor2@example.com"
	names2 := []string{"Contributor Two", "Alt Name"}
	weight2 := 0.4
	require.NoError(t, db.InsertRepoMetric(request.Id, repo.Id, gitEmail2, names2, weight2, time.Now()))

	results, err := db.FindAnalysisResults(request.Id)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFindRepoContribution(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	request := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(request, time.Now()))
	require.NoError(t, db.UpdateAnalysisRequest(request.Id, time.Now(), nil))

	gitEmail := "contributor@example.com"
	names := []string{"Contributor"}
	weight := 0.8
	require.NoError(t, db.InsertRepoMetric(request.Id, repo.Id, gitEmail, names, weight, time.Now()))

	contributions, err := db.FindRepoContribution(repo.Id)
	require.NoError(t, err)
	assert.Len(t, contributions, 1)
	assert.Equal(t, gitEmail, contributions[0].GitEmail)
	assert.Equal(t, weight, contributions[0].Weight)
	assert.Equal(t, names, contributions[0].GitNames)
}

func TestFindRepoContribution_ExcludesErrors(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	successRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(successRequest, time.Now()))
	require.NoError(t, db.UpdateAnalysisRequest(successRequest.Id, time.Now(), nil))

	gitEmail := "contributor@example.com"
	names := []string{"Contributor"}
	weight := 0.8
	require.NoError(t, db.InsertRepoMetric(successRequest.Id, repo.Id, gitEmail, names, weight, time.Now()))

	errorRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -60),
		DateTo:   time.Now().AddDate(0, 0, -30),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(errorRequest, time.Now()))
	errStr := "analysis failed"
	require.NoError(t, db.UpdateAnalysisRequest(errorRequest.Id, time.Now(), &errStr))

	require.NoError(t, db.InsertRepoMetric(errorRequest.Id, repo.Id, "error@example.com", names, 0.5, time.Now()))

	contributions, err := db.FindRepoContribution(repo.Id)
	require.NoError(t, err)
	assert.Len(t, contributions, 1)
	assert.Equal(t, gitEmail, contributions[0].GitEmail)
}

func TestFindRepoContributors(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	request := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(request, time.Now()))
	require.NoError(t, db.UpdateAnalysisRequest(request.Id, time.Now(), nil))

	gitEmail1 := "contributor1@example.com"
	gitEmail2 := "contributor2@example.com"
	names := []string{"Contributor"}
	weight := 0.5

	require.NoError(t, db.InsertRepoMetric(request.Id, repo.Id, gitEmail1, names, weight, time.Now()))
	require.NoError(t, db.InsertRepoMetric(request.Id, repo.Id, gitEmail2, names, weight, time.Now()))
	require.NoError(t, db.InsertRepoMetric(request.Id, repo.Id, gitEmail1, names, weight, time.Now()))

	count, err := db.FindRepoContributors(repo.Id)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestFindRepoContributors_ExcludesErrors(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	successRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(successRequest, time.Now()))
	require.NoError(t, db.UpdateAnalysisRequest(successRequest.Id, time.Now(), nil))

	gitEmail := "contributor@example.com"
	names := []string{"Contributor"}
	weight := 0.8
	require.NoError(t, db.InsertRepoMetric(successRequest.Id, repo.Id, gitEmail, names, weight, time.Now()))

	errorRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -60),
		DateTo:   time.Now().AddDate(0, 0, -30),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(errorRequest, time.Now()))
	errStr := "analysis failed"
	require.NoError(t, db.UpdateAnalysisRequest(errorRequest.Id, time.Now(), &errStr))

	require.NoError(t, db.InsertRepoMetric(errorRequest.Id, repo.Id, "error@example.com", names, 0.5, time.Now()))

	count, err := db.FindRepoContributors(repo.Id)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}