package types

type Opts struct {
	Port      int
	HS256     string
	Env       string
	DBPath    string
	DBDriver  string
	DBScripts string
	Admins    string
}
