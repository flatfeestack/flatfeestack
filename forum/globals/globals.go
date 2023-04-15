package globals

import (
	"database/sql"
	"forum/types"
	"forum/utils"
)

var (
	DB     *sql.DB
	OPTS   *types.Opts
	JwtKey []byte
	ADMINS []string
	KM     = utils.KeyedMutex{}
)
