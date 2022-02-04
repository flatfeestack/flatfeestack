package main

import (
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=testdb",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	resource.Expire(60 * 1000) //1min
	port := resource.GetPort("5432/tcp")
	log.Printf("DB port: %v\n", port)
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", fmt.Sprintf("postgresql://postgres:password@localhost:%s/testdb?sslmode=disable", port))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	err = db.Close()
	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func runSQLFile(files ...string) error {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			_, err = db.Exec(string(fileBytes))
			if err != nil {
				return fmt.Errorf("[%v] %v", string(fileBytes), err)
			}
		}
	}
	return nil
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
	runSQLFile("init_func.sql")
	if err != nil {
		log.Fatalf("Could not run init_func.sql scripts: %s", err)
	}
}
func teardown() {
	err := runSQL("drop_test.sql")
	if err != nil {
		log.Fatalf("Could not run drop_test.sql: %s", err)
	}
}
