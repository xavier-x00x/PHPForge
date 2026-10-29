package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ServerConfig represents the global server configuration
type ServerConfig struct {
	APIPort    int    `json:"api_port"`
	DataDir    string `json:"data_dir"`
	PHPWorkers int    `json:"php_workers"`
	APIKey     string `json:"api_key"`
}

// SiteConfig represents the configuration for a single site/vhost
type SiteConfig struct {
	Name         string   `json:"name"`
	Domain       string   `json:"domain"`
	Port         int      `json:"port"`
	DocumentRoot string   `json:"document_root"`
	PHPVersion   string   `json:"php_version"`
}

// SitesFile represents the structure of sites.json
type SitesFile struct {
	Sites []SiteConfig `json:"sites"`
}

func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// return default if not found
		exePath, _ := os.Executable()
		baseDir := filepath.Join(filepath.Dir(exePath), "dist")
		return &ServerConfig{
			APIPort:    8081,
			DataDir:    baseDir,
			PHPWorkers: 16,
			APIKey:     "phforge-secret-key",
		}, nil
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	if config.PHPWorkers == 0 {
		config.PHPWorkers = 16
	}
	
	if config.APIKey == "" {
		config.APIKey = "phforge-secret-key"
	}
	
	return &config, nil
}

// LoadSitesConfig reads the sites configuration from a JSON file
func LoadSitesConfig(path string) (*SitesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// return empty sites list if not found
		return &SitesFile{Sites: []SiteConfig{}}, nil
	}

	var sf SitesFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, err
	}
	
	// Auto-heal broken absolute paths for portability
	// The path variable is usually .../dist/config/sites.json
	baseDir := filepath.Dir(filepath.Dir(path)) // This gets the "dist" directory
	
	for i, site := range sf.Sites {
		if filepath.IsAbs(site.DocumentRoot) {
			if _, err := os.Stat(site.DocumentRoot); os.IsNotExist(err) {
				// Path doesn't exist, check if it was originally inside dist/sites/
				slashPath := filepath.ToSlash(site.DocumentRoot)
				idx := strings.Index(strings.ToLower(slashPath), "/dist/sites/")
				if idx != -1 {
					relPath := slashPath[idx+12:] // everything after /dist/sites/
					newAbs := filepath.Join(baseDir, "sites", relPath)
					if _, err := os.Stat(newAbs); err == nil {
						sf.Sites[i].DocumentRoot = newAbs
					}
				}
			}
		}
	}
	
	return &sf, nil
}

// SaveServerConfig writes the server configuration to a JSON file
func SaveServerConfig(path string, config ServerConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(path), 0755)
	
	return os.WriteFile(path, data, 0644)
}

// SaveSitesConfig writes the sites configuration to a JSON file
func SaveSitesConfig(path string, config SitesFile) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(path), 0755)
	
	return os.WriteFile(path, data, 0644)
}
