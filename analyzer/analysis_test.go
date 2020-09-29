package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"archive/zip"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

func TestAnalyzeRepositoryFromRepository(t *testing.T) {
	_, _ = Unzip("test-repository.zip", "test-repository")

	repo, _ := git.PlainOpen("./test-repository")

	var defaultTime time.Time

	contributions, err := analyzeRepositoryFromRepository(repo, defaultTime, defaultTime, "master")

	expectedContributions := make(map[Contributor]Contribution)

	expectedContributions[Contributor{
		Name:  "Claude Muller",
		Email: "claude@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Claude Muller",
			Email: "claude@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 20297,
			Deletion: 12134,
		},
		Merges:      11,
		Commits:     143,
	}
	expectedContributions[Contributor{
		Name:  "Claude Muller",
		Email: "37138571+claudemiller@users.noreply.github.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Claude Muller",
			Email: "37138571+claudemiller@users.noreply.github.com",
		},
		Changes:     CommitChange{
			Addition: 839,
			Deletion: 652,
		},
		Merges:      5,
		Commits:     4,
	}
	expectedContributions[Contributor{
		Name:  "Guil. Sperb Machado",
		Email: "guil@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Guil. Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 948,
			Deletion: 398,
		},
		Merges:      18,
		Commits:     0,
	}
	expectedContributions[Contributor{
		Name:  "Guil. Sperb Machado",
		Email: "gsm@machados.org",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Guil. Sperb Machado",
			Email: "gsm@machados.org",
		},
		Changes:     CommitChange{
			Addition: 0,
			Deletion: 0,
		},
		Merges:      1,
		Commits:     0,
	}
	expectedContributions[Contributor{
		Name:  "Sebastian Stephan",
		Email: "sebastian-stephan@users.noreply.github.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Sebastian Stephan",
			Email: "sebastian-stephan@users.noreply.github.com",
		},
		Changes:     CommitChange{
			Addition: 1,
			Deletion: 1,
		},
		Merges:      0,
		Commits:     1,
	}
	expectedContributions[Contributor{
		Name:  "Nikita Andrejevs",
		Email: "nimmortalz@gmail.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Nikita Andrejevs",
			Email: "nimmortalz@gmail.com",
		},
		Changes:     CommitChange{
			Addition: 3720,
			Deletion: 1303,
		},
		Merges:      4,
		Commits:     7,
	}
	expectedContributions[Contributor{
		Name:  "Freddy Tuxworth",
		Email: "freddytuxworth@gmail.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Freddy Tuxworth",
			Email: "freddytuxworth@gmail.com",
		},
		Changes:     CommitChange{
			Addition: 47,
			Deletion: 0,
		},
		Merges:      0,
		Commits:     1,
	}
	expectedContributions[Contributor{
		Name:  "Guilherme Sperb Machado",
		Email: "guil@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Guilherme Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 39585,
			Deletion: 14353,
		},
		Merges:      18,
		Commits:     188,
	}
	expectedContributions[Contributor{
		Name:  "Krain Chen",
		Email: "chenquanyu@ngd.neo.org",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Krain Chen",
			Email: "chenquanyu@ngd.neo.org",
		},
		Changes:     CommitChange{
			Addition: 8,
			Deletion: 5,
		},
		Merges:      0,
		Commits:     1,
	}
	expectedContributions[Contributor{
		Name:  "Nikita Andrejevs",
		Email: "nikita.andrejevs@knowledgeprice.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Nikita Andrejevs",
			Email: "nikita.andrejevs@knowledgeprice.com",
		},
		Changes:     CommitChange{
			Addition: 273,
			Deletion: 55,
		},
		Merges:      2,
		Commits:     0,
	}
	expectedContributions[Contributor{
		Name:  "施鹏",
		Email: "shipeng@aladingbank.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "施鹏",
			Email: "shipeng@aladingbank.com",
		},
		Changes:     CommitChange{
			Addition: 4,
			Deletion: 2,
		},
		Merges:      0,
		Commits:     1,
	}
	expectedContributions[Contributor{
		Name:  "claudemiller",
		Email: "claude@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "claudemiller",
			Email: "claude@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 314,
			Deletion: 257,
		},
		Merges:      0,
		Commits:     1,
	}
	expectedContributions[Contributor{
		Name:  "Krain Chen",
		Email: "ssssu8@qq.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Krain Chen",
			Email: "ssssu8@qq.com",
		},
		Changes:     CommitChange{
			Addition: 60,
			Deletion: 3,
		},
		Merges:      1,
		Commits:     0,
	}

	assert.Equal(t, expectedContributions, contributions)
	assert.Equal(t, nil, err)

	_ = os.RemoveAll("./test-repository")
}

func TestAnalyzeRepositoryFromRepository_DateRange(t *testing.T) {
	_, _ = Unzip("test-repository.zip", "test-repository")

	repo, _ := git.PlainOpen("./test-repository")

	startDate := parseRFC3339WithoutError("2019-02-01T12:00:00Z")
	endDate := parseRFC3339WithoutError("2019-04-30T12:00:00Z")

	contributions, err := analyzeRepositoryFromRepository(repo, startDate, endDate, "master")

	expectedContributions := make(map[Contributor]Contribution)

	expectedContributions[Contributor{
		Name:  "Nikita Andrejevs",
		Email: "nimmortalz@gmail.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Nikita Andrejevs",
			Email: "nimmortalz@gmail.com",
		},
		Changes:     CommitChange{
			Addition: 474,
			Deletion: 131,
		},
		Merges:      0,
		Commits:     3,
	}
	expectedContributions[Contributor{
		Name:  "Freddy Tuxworth",
		Email: "freddytuxworth@gmail.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Freddy Tuxworth",
			Email: "freddytuxworth@gmail.com",
		},
		Changes:     CommitChange{
			Addition: 47,
			Deletion: 0,
		},
		Merges:      0,
		Commits:     1,
	}
	expectedContributions[Contributor{
		Name:  "Guilherme Sperb Machado",
		Email: "guil@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Guilherme Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 3634,
			Deletion: 457,
		},
		Merges:      0,
		Commits:     34,
	}


	assert.Equal(t, expectedContributions, contributions)
	assert.Equal(t, nil, err)

	_ = os.RemoveAll("./test-repository")
}

func TestWeightContributions(t *testing.T) {

	inputContributions := make(map[Contributor]Contribution)

	inputContributions[Contributor{
		Name:  "Claude Muller",
		Email: "claude@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Claude Muller",
			Email: "claude@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 10245,
			Deletion: 6405,
		},
		Merges:      6,
		Commits:     64,
	}
	inputContributions[Contributor{
		Name:  "Claude Muller",
		Email: "37138571+claudemiller@users.noreply.github.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Claude Muller",
			Email: "37138571+claudemiller@users.noreply.github.com",
		},
		Changes:     CommitChange{
			Addition: 668,
			Deletion: 372,
		},
		Merges:      3,
		Commits:     1,
	}
	inputContributions[Contributor{
		Name:  "Guil. Sperb Machado",
		Email: "guil@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Guil. Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 90,
			Deletion: 25,
		},
		Merges:      3,
		Commits:     0,
	}
	inputContributions[Contributor{
		Name:  "Nikita Andrejevs",
		Email: "nimmortalz@gmail.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Nikita Andrejevs",
			Email: "nimmortalz@gmail.com",
		},
		Changes:     CommitChange{
			Addition: 2116,
			Deletion: 729,
		},
		Merges:      2,
		Commits:     2,
	}
	inputContributions[Contributor{
		Name:  "Guilherme Sperb Machado",
		Email: "guil@axlabs.com",
	}] = Contribution{
		Contributor: Contributor{
			Name:  "Guilherme Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Changes:     CommitChange{
			Addition: 2571,
			Deletion: 1093,
		},
		Merges:      8,
		Commits:     35,
	}

	outputScore, err := weightContributions(inputContributions)

	expectedOutput := make(map[Contributor]FlatFeeWeight)
	expectedOutput[Contributor{
		Name:  "Claude Muller",
		Email: "claude@axlabs.com",
	}] = FlatFeeWeight{
		Contributor: Contributor{
			Name:  "Claude Muller",
			Email: "claude@axlabs.com",
		},
		Weight: 0.6453751874866082,
	}
	expectedOutput[Contributor{
		Name:  "Claude Muller",
		Email: "37138571+claudemiller@users.noreply.github.com",
	}] = FlatFeeWeight{
		Contributor: Contributor{
			Name:  "Claude Muller",
			Email: "37138571+claudemiller@users.noreply.github.com",
		},
		Weight: 0.03514431962342826,
	}
	expectedOutput[Contributor{
		Name:  "Guil. Sperb Machado",
		Email: "guil@axlabs.com",
	}] = FlatFeeWeight{
		Contributor: Contributor{
			Name:  "Guil. Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Weight: 0.007351913638821717,
	}
	expectedOutput[Contributor{
		Name:  "Nikita Andrejevs",
		Email: "nimmortalz@gmail.com",
	}] = FlatFeeWeight{
		Contributor: Contributor{
			Name:  "Nikita Andrejevs",
			Email: "nimmortalz@gmail.com",
		},
		Weight: 0.09139425415191431,
	}
	expectedOutput[Contributor{
		Name:  "Guilherme Sperb Machado",
		Email: "guil@axlabs.com",
	}] = FlatFeeWeight{
		Contributor: Contributor{
			Name:  "Guilherme Sperb Machado",
			Email: "guil@axlabs.com",
		},
		Weight: 0.22073432509922764,
	}

	sumOfScores := 0.0

	for k, v := range outputScore {
		expected, found := expectedOutput[k]
		sumOfScores += v.Weight
		assert.Equal(t, true, found)
		assert.Equal(t, fmt.Sprintf("%.12f", expected.Weight), fmt.Sprintf("%.12f", v.Weight))
	}

	for k, _ := range expectedOutput {
		_, found := outputScore[k]
		assert.Equal(t, true, found)
	}
	assert.Equal(t, 1.0, sumOfScores)
	assert.Equal(t, nil, err)
}