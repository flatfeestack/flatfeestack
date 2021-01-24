package main

import (
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"io/ioutil"
	"log"
	"os"
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
	tmpName, err := ioutil.TempDir(os.TempDir(), "postgresql")
	if err != nil {
		log.Fatal(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=testdb",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		config.Mounts = []docker.HostMount{docker.HostMount{
			Target: "/var/lib/postgresql/data",
			Source: tmpName,
			Type:   "bind",
		}}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	resource.Expire(3600) //1h
	port := resource.GetPort("5432/tcp")
	log.Printf("DB port: %v\n", port)
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", fmt.Sprintf("postgresql://postgres:password@localhost:%s/testdb?sslmode=disable", port))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.RemoveAll(tmpName)
	os.Exit(code)
}

func runSQL(db *sql.DB, files ...string) error {
	for _, file := range files {
		//this will stringPointer or alter tables
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			file, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			requests := strings.Split(string(file), ";")
			for _, request := range requests {
				lines := strings.Split(request, "\n")
				cleanRequest := ""
				for _, line := range lines {
					line = strings.Replace(line, "\t", "", -1)
					if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "--") && len(line) > 0 {
						cleanRequest += strings.TrimSpace(line) + " "
					}
				}

				_, err := db.Exec(cleanRequest)
				if err != nil {
					return fmt.Errorf("[%v] %v", request, err)
				}
			}
		} else {
			log.Printf("ignoring file %v (%v)", file, err)
		}
	}
	return nil
}

func setup() {
	err := runSQL(db, "init.sql", "sched.sql")
	if err != nil {
		log.Fatalf("Could run init scripts: %s", err)
	}
}
func teardown() {
	err := runSQL(db, "drop_test.sql")
	if err != nil {
		log.Fatalf("Could run drop_test.sql: %s", err)
	}
}

func stringPointer(s string) *string {
	return &s
}
