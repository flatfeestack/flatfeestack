package db

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdateTrustedRepoTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
		UnTrustAt: &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t2,
		UnTrustAt: &t2,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "to trust, we must set the trust_at, but not the untrust_at: event.TrustAt: 0001-01-01 00:00:02 +0000 UTC, event.UnTrustAt: 0001-01-01 00:00:02 +0000 UTC")

}

func TestUnTrustedTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t2,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t3,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr3)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "we want to untrust, but we already untrusted it: unTrustAt: 0001-01-01 00:00:02 +0000 UTC")

}

func TestUnTrustWrong(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t1,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "we want to untrust, but we are currently not trusting this repo")
}

func TestTrustWrongOrder(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t2,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t1,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "we want to untrust, but the untrust date is before the trust date: trustAt: 0001-01-01 00:00:02 +0000 UTC, event.UnTrustAt: 0001-01-01 00:00:01 +0000 UTC")

}

func TestTrustWrongOrderActive(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t2,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t4,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t3,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr3)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "we want to untrust, but we already untrusted it: unTrustAt: 0001-01-01 00:00:04 +0000 UTC")

}

func TestTrustCorrect(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)
	r2 := insertTestRepoGitUrl(t, "git-url2")

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t2,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		TrustAt:   &t3,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr3)
	assert.Nil(t, err)

	rs, err := FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, r2.Id, rs[0].Id)

	tr4 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Inactive,
		UnTrustAt: &t4,
	}
	err = InsertOrUpdateTrustRepo(&tr4)
	assert.Nil(t, err)

	rs, err = FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestTwoTrustedRepos(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		TrustAt:   &t2,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)

	rs, err := FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(rs))
}

func TestGeneratePlaceholders(t *testing.T) {
	t.Run("ZeroPlaceholders", func(t *testing.T) {
		result := GeneratePlaceholders(0)
		expected := ""
		assert.Equal(t, expected, result, "Expected empty string for 0 placeholders")
	})

	t.Run("OnePlaceholder", func(t *testing.T) {
		result := GeneratePlaceholders(1)
		expected := "$1"
		assert.Equal(t, expected, result, "Expected '$1' for 1 placeholder")
	})

	t.Run("MultiplePlaceholders", func(t *testing.T) {
		result := GeneratePlaceholders(5)
		expected := "$1, $2, $3, $4, $5"
		assert.Equal(t, expected, result, "Expected placeholders for 5 inputs")
	})

	t.Run("LargeNumberPlaceholders", func(t *testing.T) {
		result := GeneratePlaceholders(100)
		parts := strings.Split(result, ", ")
		assert.Equal(t, 100, len(parts), "Expected 100 placeholders")
		assert.Equal(t, "$1", parts[0], "First placeholder should be $1")
		assert.Equal(t, "$100", parts[99], "Last placeholder should be $100")
	})
}

func TestConvertToInterfaceSlice(t *testing.T) {
	t.Run("EmptySlice", func(t *testing.T) {
		input := []int{}
		result := ConvertToInterfaceSlice(input)
		assert.Equal(t, 0, len(result), "Expected empty interface slice for empty input slice")
	})

	t.Run("SingleElementSlice", func(t *testing.T) {
		input := []string{"test"}
		result := ConvertToInterfaceSlice(input)
		assert.Equal(t, 1, len(result), "Expected one element in interface slice")
		assert.Equal(t, "test", result[0], "Element should match input")
	})

	t.Run("MultipleElementSlice", func(t *testing.T) {
		input := []int{1, 2, 3}
		result := ConvertToInterfaceSlice(input)
		expected := []interface{}{1, 2, 3}
		assert.Equal(t, expected, result, "Expected interface slice to match input slice")
	})

	t.Run("LargeSlice", func(t *testing.T) {
		input := make([]int, 100)
		for i := 0; i < 100; i++ {
			input[i] = i
		}
		result := ConvertToInterfaceSlice(input)
		assert.Equal(t, 100, len(result), "Expected 100 elements in interface slice")
		assert.Equal(t, 0, result[0], "First element should match input")
		assert.Equal(t, 99, result[99], "Last element should match input")
	})
}
