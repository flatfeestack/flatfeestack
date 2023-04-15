package database

import (
	"database/sql"
	"fmt"
	"forum/globals"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

// InitDb stringPointer connection with postgres db
func InitDb() (*sql.DB, error) {
	// Open the connection
	db, err := sql.Open(globals.OPTS.DBDriver, globals.OPTS.DBPath)
	if err != nil {
		return nil, err
	}

	//we wait for ten seconds to connect
	err = db.Ping()
	now := time.Now()
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now().UTC()) {
		time.Sleep(time.Second)
		err = db.Ping()
	}
	if err != nil {
		return nil, err
	}

	//this will create or alter tables
	//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	for _, file := range strings.Split(globals.OPTS.DBScripts, ":") {
		if file == "" {
			continue
		}
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, err
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
						return nil, fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		} else {
			log.Infof("ignoring file [%v] (%v)", file, err)
		}
	}

	log.Infof("Successfully connected to Database!")
	return db, nil
}
