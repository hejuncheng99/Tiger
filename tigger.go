package TIGER

import "database/sql"

// Config Tiger Config
type Config struct {
	DNS        string
	DriverName string
}

func Open(config *Config) (*sql.DB, error) {
	open, err := sql.Open(config.DriverName, config.DNS)
	if err != nil {
		return nil, err
	}

	return open, nil
}
