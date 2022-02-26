package main

import (
	"archive/zip"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestRemoteRepo(t *testing.T) {
	opts = &Opts{}
	opts.GitBasePath = "/tmp"
	//r, err := cloneOrUpdateRepository("git@github.com:flatfeestack/flatfeestack-test-itself.git", "master")
	//r, err := cloneOrUpdateRepository("https://github.com/torvalds/linux.git", "master")
	r, err := cloneOrUpdateRepository("https://github.com/neow3j/neow3j.git", "master-3.x")
	assert.Nil(t, err)
	var defaultTime time.Time
	c, err := analyzeRepositoryFromRepository(r, defaultTime, defaultTime)
	assert.Nil(t, err)
	f, err := weightContributions(c)
	assert.Nil(t, err)
	sort.Slice(f, func(i, j int) bool {
		return f[i].Weight > f[j].Weight
	})
	for _, v := range f {
		fmt.Printf("out: %v\n", v)
	}
}

// Helpers

func RoundToDecimals(f float64, decimals int) float64 {
	return math.Round(f*float64(10)*float64(decimals)) / (float64(10) * float64(decimals))
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// Tests

// analyzeRepositoryFromString

/*
	This adds just the part of cloning/updating the repository
	before the analysis that is already tested. Since we don't test
	go-git we won't test that.
*/

// analyzeRepositoryFromRepository

/*func TestAnalyzeRepositoryFromRepository(t *testing.T) {
	_, _ = Unzip("test-repository.zip", "test-repository")

	repo, _ := git.PlainOpen("./test-repository")

	var defaultTime time.Time

	contributions, err := analyzeRepositoryFromRepository(repo, defaultTime, defaultTime)

	expectedContributions := make(map[string]Contribution)

	expectedContributions["claude@axlabs.com"] = Contribution{
		Names:    []string{"Claude Muller", "claudemiller"},
		Addition: 15419 + 263,
		Deletion: 7256 + 206,
		Merges:   11 + 0,
		Commits:  143 + 1,
	}
	expectedContributions["37138571+claudemiller@users.noreply.github.com"] = Contribution{
		Names:    []string{"Claude Muller"},
		Addition: 674,
		Deletion: 488,
		Merges:   5,
		Commits:  4,
	}
	expectedContributions["guil@axlabs.com"] = Contribution{
		Names:    []string{"Guil. Sperb Machado", "Guilherme Sperb Machado"},
		Addition: 864 + 29253,
		Deletion: 315 + 4021,
		Merges:   18 + 18,
		Commits:  0 + 188,
	}
	expectedContributions["gsm@machados.org"] = Contribution{
		Names:    []string{"Guil. Sperb Machado"},
		Addition: 0,
		Deletion: 0,
		Merges:   1,
		Commits:  0,
	}
	expectedContributions["sebastian-stephan@users.noreply.github.com"] = Contribution{
		Names:    []string{"Sebastian Stephan"},
		Addition: 1,
		Deletion: 1,
		Merges:   0,
		Commits:  1,
	}
	expectedContributions["nimmortalz@gmail.com"] = Contribution{
		Names:    []string{"Nikita Andrejevs"},
		Addition: 3433,
		Deletion: 1016,
		Merges:   4,
		Commits:  7,
	}
	expectedContributions["freddytuxworth@gmail.com"] = Contribution{
		Names:    []string{"Freddy Tuxworth"},
		Addition: 47,
		Deletion: 0,
		Merges:   0,
		Commits:  1,
	}
	expectedContributions["chenquanyu@ngd.neo.org"] = Contribution{
		Names:    []string{"Krain Chen"},
		Addition: 8,
		Deletion: 5,
		Merges:   0,
		Commits:  1,
	}
	expectedContributions["nikita.andrejevs@knowledgeprice.com"] = Contribution{
		Names:    []string{"Nikita Andrejevs"},
		Addition: 242,
		Deletion: 24,
		Merges:   2,
		Commits:  0,
	}
	expectedContributions["shipeng@aladingbank.com"] = Contribution{
		Names:    []string{"施鹏"},
		Addition: 4,
		Deletion: 2,
		Merges:   0,
		Commits:  1,
	}

	expectedContributions["ssssu8@qq.com"] = Contribution{
		Names:    []string{"Krain Chen"},
		Addition: 60,
		Deletion: 3,
		Merges:   1,
		Commits:  0,
	}

	assert.Equal(t, expectedContributions, contributions)
	assert.Equal(t, nil, err)

	_ = os.RemoveAll("./test-repository")
}

func TestAnalyzeRepositoryFromRepository_DateRange(t *testing.T) {
	_, _ = Unzip("test-repository.zip", "test-repository")

	repo, _ := git.PlainOpen("./test-repository")

	startDate, err := time.Parse(time.RFC3339, "2019-02-01T12:00:00Z")
	endDate, err := time.Parse(time.RFC3339, "2019-04-30T12:00:00Z")

	contributions, err := analyzeRepositoryFromRepository(repo, startDate, endDate)

	expectedContributions := make(map[string]Contribution)

	expectedContributions["nimmortalz@gmail.com"] = Contribution{
		Names:    []string{"Nikita Andrejevs"},
		Addition: 474,
		Deletion: 131,
		Merges:   0,
		Commits:  3,
	}
	expectedContributions["freddytuxworth@gmail.com"] = Contribution{
		Names:    []string{"Freddy Tuxworth"},
		Addition: 47,
		Deletion: 0,
		Merges:   0,
		Commits:  1,
	}
	expectedContributions["guil@axlabs.com"] = Contribution{
		Names:    []string{"Guilherme Sperb Machado"},
		Addition: 3527,
		Deletion: 350,
		Merges:   0,
		Commits:  34,
	}

	assert.Equal(t, expectedContributions, contributions)
	assert.Equal(t, nil, err)

	_ = os.RemoveAll("./test-repository")
}

func TestAnalyzeRepositoryFromRepository_EmptyRepo(t *testing.T) {

	storer := memory.NewStorage()

	repo, _ := git.Init(storer, nil)

	var defaultTime time.Time

	contributions, err := analyzeRepositoryFromRepository(repo, defaultTime, defaultTime)

	expectedContributions := make(map[string]Contribution)

	// Expect error since git logging an empty repo returns an error
	assert.NotEqual(t, nil, err)
	assert.Equal(t, expectedContributions, contributions)
}*/

// weightContributions

func TestWeightContributions(t *testing.T) {

	inputContributions := make(map[string]Contribution)

	inputContributions["claude@axlabs.com"] = Contribution{
		Names:    []string{"Claude Muller"},
		Addition: 10245,
		Deletion: 6405,
		Merges:   6,
		Commits:  64,
	}
	inputContributions["37138571+claudemiller@users.noreply.github.com"] = Contribution{
		Names:    []string{"Claude Muller"},
		Addition: 668,
		Deletion: 372,
		Merges:   3,
		Commits:  1,
	}
	inputContributions["guil@axlabs.com"] = Contribution{
		Names:    []string{"Guil. Sperb Machado"},
		Addition: 90,
		Deletion: 25,
		Merges:   3,
		Commits:  0,
	}
	inputContributions["nimmortalz@gmail.com"] = Contribution{
		Names:    []string{"Nikita Andrejevs"},
		Addition: 2116,
		Deletion: 729,
		Merges:   2,
		Commits:  2,
	}
	inputContributions["guil@axlabs.com"] = Contribution{
		Names:    []string{"Guilherme Sperb Machado"},
		Addition: 2571,
		Deletion: 1093,
		Merges:   8,
		Commits:  35,
	}

	outputScore, err := weightContributions(inputContributions)

	expectedOutput := make(map[string]FlatFeeWeight)
	expectedOutput["claude@axlabs.com"] = FlatFeeWeight{
		Names:  []string{"Claude Muller"},
		Weight: 0.642448223882,
	}
	expectedOutput["37138571+claudemiller@users.noreply.github.com"] = FlatFeeWeight{
		Names:  []string{"Claude Muller"},
		Weight: 0.032936082828,
	}
	expectedOutput["guil@axlabs.com"] = FlatFeeWeight{
		Names:  []string{"Guil. Sperb Machado"},
		Weight: 0.008049671622095021,
	}
	expectedOutput["nimmortalz@gmail.com"] = FlatFeeWeight{
		Names:  []string{"Nikita Andrejevs"},
		Weight: 0.080929875974,
	}
	expectedOutput["guil@axlabs.com"] = FlatFeeWeight{
		Names:  []string{"Guilherme Sperb Machado"},
		Weight: 0.243685817316,
	}

	sumOfScores := 0.0

	for _, v := range outputScore {
		expected, found := expectedOutput[v.Email]
		sumOfScores += v.Weight
		assert.Equal(t, true, found)
		assert.Equal(t, fmt.Sprintf("%.12f", expected.Weight), fmt.Sprintf("%.12f", v.Weight))
	}

	assert.Equal(t, len(outputScore), len(expectedOutput))
	assert.Equal(t, 1.0, RoundToDecimals(sumOfScores, 12))
	assert.Equal(t, nil, err)
}

func TestWeightContributions_OneInput(t *testing.T) {

	inputContributions := make(map[string]Contribution)

	inputContributions["claude@axlabs.com"] = Contribution{
		Names:    []string{"Claude Muller"},
		Addition: 10245,
		Deletion: 6405,
		Merges:   6,
		Commits:  64,
	}

	outputScore, err := weightContributions(inputContributions)

	expectedOutput := make(map[string]FlatFeeWeight)
	expectedOutput["claude@axlabs.com"] = FlatFeeWeight{
		Names:  []string{"Claude Muller"},
		Weight: 1.0,
	}

	sumOfScores := 0.0

	for _, v := range outputScore {
		expected, found := expectedOutput[v.Email]
		sumOfScores += v.Weight
		assert.Equal(t, true, found)
		assert.Equal(t, fmt.Sprintf("%.12f", expected.Weight), fmt.Sprintf("%.12f", v.Weight))
	}

	assert.Equal(t, len(outputScore), len(expectedOutput))
	assert.Equal(t, 1.0, RoundToDecimals(sumOfScores, 12))
	assert.Equal(t, nil, err)
}

func TestWeightContributions_NoInput(t *testing.T) {

	inputContributions := make(map[string]Contribution)

	outputScore, err := weightContributions(inputContributions)

	expectedOutput := make(map[string]FlatFeeWeight)

	for _, v := range outputScore {
		expected, found := expectedOutput[v.Email]
		assert.Equal(t, true, found)
		assert.Equal(t, fmt.Sprintf("%.12f", expected.Weight), fmt.Sprintf("%.12f", v.Weight))
	}

	assert.Equal(t, len(outputScore), len(expectedOutput))
	assert.Equal(t, nil, err)
}
