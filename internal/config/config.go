package config

import (
	"encoding/json"
	"os"
)

type Config struct {
    DBMS         string `json:"dbms"`
    Connection   string `json:"connection"`
    SnapshotDir  string `json:"snapshotDir"`
    MigrationDir string `json:"migrationDir"`
}

func InitConfig(cfg Config) error {
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(".schema_manager/config.json", data, 0644)
}

func LoadConfig() (Config, error) {
    data, err := os.ReadFile(".schema_manager/config.json")
    if err != nil {
        return Config{}, err
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return Config{}, err
    }
    return cfg, nil
}