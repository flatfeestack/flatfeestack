package main

import (
	"archive/zip"
	"fmt"
	git "github.com/libgit2/git2go/v33"
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
	start := time.Now()
	opts = &Opts{}
	opts.GitBasePath = "/tmp"
	//r, err := cloneOrUpdateRepository("git@github.com:flatfeestack/flatfeestack-test-itself.git")
	//r, err := cloneOrUpdateRepository("git://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git", "git://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git")
	//r, err := cloneOrUpdateRepository("git://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git")
	r, err := cloneOrUpdateRepository("https://github.com/neow3j/neow3j.git")
	assert.Nil(t, err)
	var defaultTime time.Time
	month3 := time.Now().AddDate(0, -3, 0)
	start = time.Now()
	c, err := analyzeRepositoryFromRepository(r, month3, defaultTime)
	fmt.Printf(" elpased2 %vs\n", time.Since(start).Seconds())
	assert.Nil(t, err)
	start = time.Now()
	f, err := weightContributions(c)
	assert.Nil(t, err)
	sort.Slice(f, func(i, j int) bool {
		return f[i].Weight > f[j].Weight
	})
	for _, v := range f {
		fmt.Printf("out: %v\n", v)
	}
	fmt.Printf(" elpased3 %vs\n", time.Since(start).Seconds())
}

func TestRemoteRepo2(t *testing.T) {
	opts = &Opts{}
	opts.GitBasePath = "/tmp"
	//r, err := cloneOrUpdateRepository("git@github.com:flatfeestack/flatfeestack-test-itself.git")
	//r, err := cloneOrUpdateRepository("git://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git", "git://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git")
	//r, err := cloneOrUpdateRepository("git://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git")
	r, err := cloneOrUpdateRepository("https://github.com/flatfeestack/flatfeestack-test-itself3.git")
	assert.Nil(t, err)
	var defaultTime time.Time
	month6 := time.Now().AddDate(0, -3, 0)
	c, err := analyzeRepositoryFromRepository(r, month6, defaultTime)
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

func TestAnalyzeRepositoryFromRepository(t *testing.T) {
	_, _ = unzip("test-repository.zip", "/tmp/test-repository")
	repo, _ := git.OpenRepository("/tmp/test-repository")

	var defaultTime time.Time

	contributions, err := analyzeRepositoryFromRepository(repo, defaultTime, defaultTime)

	expectedContributions := make(map[string]Contribution)

	expectedContributions["claude@axlabs.com"] = Contribution{
		Names:    []string{"Claude Muller", "claudemiller"},
		Addition: 20691,
		Deletion: 12443,
		Merges:   11,
		Commits:  155,
	}
	expectedContributions["37138571+claudemiller@users.noreply.github.com"] = Contribution{
		Names:    []string{"Claude Muller"},
		Addition: 891,
		Deletion: 690,
		Merges:   5,
		Commits:  9,
	}
	expectedContributions["guil@axlabs.com"] = Contribution{
		Names:    []string{"Guil. Sperb Machado", "Guilherme Sperb Machado"},
		Addition: 33775,
		Deletion: 14866,
		Merges:   36,
		Commits:  223,
	}
	expectedContributions["gsm@machados.org"] = Contribution{
		Names:    []string{"Guil. Sperb Machado"},
		Addition: 0,
		Deletion: 0,
		Merges:   1,
		Commits:  1,
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
		Addition: 3873,
		Deletion: 1397,
		Merges:   4,
		Commits:  11,
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
		Addition: 300,
		Deletion: 65,
		Merges:   2,
		Commits:  2,
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
		Addition: 63,
		Deletion: 5,
		Merges:   1,
		Commits:  1,
	}

	assert.Equal(t, expectedContributions, contributions)
	assert.Equal(t, nil, err)

	_ = os.RemoveAll("./test-repository")
}

func TestAnalyzeRepositoryFromRepository_DateRange(t *testing.T) {
	_, _ = unzip("test-repository.zip", "/tmp/test-repository")
	repo, _ := git.OpenRepository("/tmp/test-repository")

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
		Addition: 3640,
		Deletion: 463,
		Merges:   0,
		Commits:  34,
	}

	assert.Equal(t, expectedContributions, contributions)
	assert.Equal(t, nil, err)

	_ = os.RemoveAll("./test-repository")
}

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
		Weight: 0.638986623422,
	}
	expectedOutput["37138571+claudemiller@users.noreply.github.com"] = FlatFeeWeight{
		Names:  []string{"Claude Muller"},
		Weight: 0.031833523386,
	}
	expectedOutput["guil@axlabs.com"] = FlatFeeWeight{
		Names:  []string{"Guil. Sperb Machado"},
		Weight: 0.008049671622095021,
	}
	expectedOutput["nimmortalz@gmail.com"] = FlatFeeWeight{
		Names:  []string{"Nikita Andrejevs"},
		Weight: 0.075930168845,
	}
	expectedOutput["guil@axlabs.com"] = FlatFeeWeight{
		Names:  []string{"Guilherme Sperb Machado"},
		Weight: 0.253249684347,
	}

	sumOfScores := 0.0

	for _, v := range outputScore {
		expected, found := expectedOutput[v.Email]
		sumOfScores += v.Weight
		assert.Equal(t, true, found)
		assert.Equal(t, fmt.Sprintf("%.12f", expected.Weight), fmt.Sprintf("%.12f", v.Weight))
	}

	assert.Equal(t, len(outputScore), len(expectedOutput))
	assert.Equal(t, 1.0, roundToDecimals(sumOfScores, 12))
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
	assert.Equal(t, 1.0, roundToDecimals(sumOfScores, 12))
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

// Helpers

func roundToDecimals(f float64, decimals int) float64 {
	return math.Round(f*float64(10)*float64(decimals)) / (float64(10) * float64(decimals))
}

func unzip(src string, dest string) ([]string, error) {

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
