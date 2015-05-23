// tsadmin
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jamesrwhite/tsadmin/database"
)

type Config struct {
	Databases []database.Database `json:"databases"`
}

func Load(configPath string) (Config, error) {
	config := Config{}

	// Read in the config file
	absPath, _ := filepath.Abs(configPath)
	configFile, err := os.Open(absPath)

	if err != nil {
		return config, err
	}

	// Decode the JSON
	parser := json.NewDecoder(configFile)

	if err = parser.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}
