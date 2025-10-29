package main

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	DB *sql.DB
)

type dbRes struct {
	email        string
	refreshToken string
	emailToken   *string
	metaSystem   *string
}

func findAuthByEmail(email string) (*dbRes, error) {
	var res dbRes
	query := `SELECT meta_system, refresh_token, email_token 
              FROM auth
              WHERE email = $1`
	err := DB.QueryRow(query, email).Scan(&res.metaSystem, &res.refreshToken, &res.emailToken)

	if err != nil {
		return nil, err
	}
	res.email = email
	return &res, nil
}

func insertOrUpdateUser(email string, emailToken string, refreshToken string, now time.Time) error {
	stmt, err := DB.Prepare(`INSERT INTO auth as a (email, email_token, refresh_token, created_at) 
								   VALUES ($1, $2, $3, $4)
								   ON CONFLICT (email) DO
								     UPDATE SET email_token = $2`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO auth for %v statement failed: %v", email, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(email, emailToken, refreshToken, now)
	return handleErrEffect(res, err, "error: INSERT INTO auth", email)
}

func deleteDbUser(email string) error {
	stmt, err := DB.Prepare(`DELETE FROM auth where email = $1`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO auth for %v statement failed: %v", email, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(email, email)
	return handleErrEffect(res, err, "error: DELETE FROM auth", email)
}

func updateRefreshToken(oldRefreshToken string, newRefreshToken string) error {
	stmt, err := DB.Prepare("UPDATE auth SET refresh_token = $1 WHERE refresh_token = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE refreshTokenfor statement failed: %v", err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(newRefreshToken, oldRefreshToken)
	return handleErrNoEffect(res, err, "error: UPDATE refreshToken", "n/a")
}

func updateSystemMeta(email string, systemMeta string) error {
	stmt, err := DB.Prepare("UPDATE auth SET meta_system = $1 WHERE email=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE meta_system statement failed: %v", err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(systemMeta, email)
	return handleErrEffect(res, err, "error: UPDATE meta_system", "n/a")
}

func updateEmailToken(email string, emailToken string) error {
	stmt, err := DB.Prepare("UPDATE auth SET email_token = NULL WHERE email = $1 AND email_token = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(email, emailToken)
	return handleErrEffect(res, err, "error: UPDATE auth", email)
}

func handleErrEffect(res sql.Result, err error, info string, email string) error {
	return handleErr(res, err, info, email, true)
}

func handleErrNoEffect(res sql.Result, err error, info string, email string) error {
	return handleErr(res, err, info, email, false)
}

func handleErr(res sql.Result, err error, info string, email string, mustHaveEffect bool) error {
	if err != nil {
		return fmt.Errorf("%v query %v failed: %v", info, email, err)
	}
	nr, err := res.RowsAffected()
	if mustHaveEffect && nr == 0 {
		return fmt.Errorf("%v rows %v, affected: %v", info, nr, email)
	}
	if err != nil {
		return fmt.Errorf("%v %v err: %v", info, email, err)
	}
	return nil
}

// /////// Setup

func CloseDb() {
	err := DB.Close()
	slog.Warn("could not close the db", slog.Any("error", err))
}

func InitDb(dbDriver string, dbPath string, dbScripts string) error {
	// Open the connection
	var err error
	DB, err = sql.Open(dbDriver, dbPath)
	if err != nil {
		return err
	}

	//we wait for ten seconds to connect
	err = DB.Ping()
	now := time.Now()
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(250 * time.Millisecond)
		err = DB.Ping()
	}
	if err != nil {
		return err
	}

	files := strings.Split(dbScripts, ":")
	err = RunSQL(files...)
	if err != nil {
		return err
	}

	slog.Info("Successfully connected!")
	return nil
}

func RunSQL(files ...string) error {
	for _, file := range files {
		if file == "" {
			continue
		}
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := os.ReadFile(file)
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
					_, err := DB.Exec(request)
					if err != nil {
						return fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		} else {
			slog.Info("ignoring file %v (%v)", slog.String("file", file), slog.Any("error", err))
		}
	}
	return nil
}

func CloseAndLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		slog.Info("could not close: %v", slog.Any("error", err))
	}
}
