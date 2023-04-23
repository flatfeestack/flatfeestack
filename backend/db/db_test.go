package db

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	day1  = time.Time{}
	day11 = time.Time{}.Add(time.Duration(1) * time.Second)
	day2  = time.Time{}.Add(time.Duration(1*24) * time.Hour)
	day3  = time.Time{}.Add(time.Duration(2*24) * time.Hour)
	day4  = time.Time{}.Add(time.Duration(3*24) * time.Hour)
	day5  = time.Time{}.Add(time.Duration(4*24) * time.Hour)
)

func TestMain(m *testing.M) {
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())

	err = InitDb("sqlite3", file.Name(), "")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = db.Close()
	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	os.Exit(code)
}

func runSQL(files ...string) error {
	for _, file := range files {
		if file == "" {
			continue
		}
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			//https://stackoverflow.com/questions/12682405/strip-out-c-style-comments-from-a-byte
			re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/|(?s)--.*?\n|(?s)#.*?\n")
			newBytes := re.ReplaceAll(fileBytes, nil)

			requests := strings.Split(string(newBytes), ";")
			for _, request := range requests {
				request = strings.TrimSpace(request)
				if len(request) > 0 {
					_, err := db.Exec(request)
					if err != nil {
						return fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		} else {
			log.Printf("ignoring file %v (%v)", file, err)
		}
	}
	return nil
}

func setup() {
	err := runSQL("init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}
func teardown() {
	err := runSQL("delAll_test.sql")
	if err != nil {
		log.Fatalf("Could not run delAll_test.sql: %s", err)
	}
}
