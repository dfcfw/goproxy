package config

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
}

type Database struct {
	DSN string `json:"dsn"`
}

type Server struct {
	Addr   string            `json:"addr"`
	Static map[string]string `json:"static"`
	CAS    string            `json:"cas"`
}
