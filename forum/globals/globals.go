package globals

import (
	"database/sql"
	"forum/types"
)

var (
	DB     *sql.DB
	OPTS   *types.Opts
	JwtKey []byte
	ADMINS []string
)
