package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ServerStatus represents the response from Core Service API
type ServerStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// GetServerStatus calls the Local REST API to get the status
func (a *App) GetServerStatus() (string, error) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get("http://127.0.0.1:8081/api/v1/status")
	if err != nil {
		return "", fmt.Errorf("Core Service is not running: %v", err)
	}
	defer resp.Body.Close()

	var status ServerStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s (v%s)", status.Status, status.Version), nil
}

// RestartCoreService reloads the core service configuration
func (a *App) RestartCoreService() (string, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	// Get API Key from config to send in header
	cfg, _ := a.GetServerConfig()
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = "phforge-secret-key"
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:8081/api/v1/reload", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Core Service is not responding: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Core Service returned error: %d", resp.StatusCode)
	}

	return "Core Service reloaded successfully", nil
}

// OpenSiteCLI opens a Windows Command Prompt set to the site's folder and configured PHP version
func (a *App) OpenSiteCLI(documentRoot string, phpVersion string) error {
	phpDirName := "php-" + phpVersion
	if strings.HasPrefix(phpVersion, "php-") {
		phpDirName = phpVersion
	}
	phpDir := filepath.Join(a.getDistDir(), "php-versions", phpDirName)
	if _, err := os.Stat(phpDir); os.IsNotExist(err) {
		phpDir = filepath.Join(a.getDistDir(), "php-versions", phpVersion)
	}

	composerDir := filepath.Join(a.getDistDir(), "composer")

	title := fmt.Sprintf("PHPForge CLI (PHP %s)", phpVersion)
	cmdStr := fmt.Sprintf(`set PATH=%s;%s;%%PATH%% & cd /d "%s"`, phpDir, composerDir, documentRoot)
	cmd := exec.Command("cmd.exe", "/c", "start", title, "cmd.exe", "/K", cmdStr)
	return cmd.Start()
}

// SelectDirectory opens a native Windows folder picker
func (a *App) SelectDirectory(title string) (string, error) {
	if title == "" {
		title = "Select Project Document Root"
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

// OpenBrowser opens a URL in the user's default web browser
func (a *App) OpenBrowser(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}

// OpenPathInExplorer opens a folder in Windows File Explorer
func (a *App) OpenPathInExplorer(targetPath string) error {
	if _, err := os.Stat(targetPath); err != nil {
		return fmt.Errorf("directory not found: %s", targetPath)
	}
	cmd := exec.Command("explorer.exe", targetPath)
	return cmd.Start()
}

// OpenPhpIni opens php.ini in Notepad (auto-creates from development ini if not found)
func (a *App) OpenPhpIni(version string) error {
	iniPath := a.getPHPIniPath(version)
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return fmt.Errorf("php.ini could not be found or initialized for PHP %s", version)
	}
	cmd := exec.Command("notepad.exe", iniPath)
	return cmd.Start()
}

// CheckPortAvailable checks if a TCP port is currently free on localhost
func (a *App) CheckPortAvailable(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	l.Close()
	return true
}

// GetServiceInfo checks if the Windows Service PHForgeCoreSvc is installed and running
func (a *App) GetServiceInfo() (map[string]interface{}, error) {
	cmd := exec.Command("sc.exe", "query", "PHForgeCoreSvc")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	out, err := cmd.CombinedOutput()
	output := string(out)

	installed := !strings.Contains(output, "does not exist") && !strings.Contains(output, "1060") && err == nil
	running := strings.Contains(output, "RUNNING")

	return map[string]interface{}{
		"installed": installed,
		"running":   running,
		"raw":       strings.TrimSpace(output),
	}, nil
}

// PHPVersionInfo holds information about a detected PHP installation
type PHPVersionInfo struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Status  string `json:"status"`
}

// getDistDir determines the path to the dist folder
func (a *App) getDistDir() string {
	exePath, _ := os.Executable()
	baseDir := filepath.Join(filepath.Dir(exePath), "dist")

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		baseDir = filepath.Join(cwd, "..", "..", "dist")
	}
	return baseDir
}

// getPHPIniPath locates or auto-initializes the php.ini file for a given PHP version
func (a *App) getPHPIniPath(version string) string {
	baseDir := a.getDistDir()
	
	// Candidates to check: "php-8.4.22" or "8.4.22"
	cleanVer := strings.TrimPrefix(version, "php-")
	dirsToCheck := []string{
		filepath.Join(baseDir, "php-versions", "php-"+cleanVer),
		filepath.Join(baseDir, "php-versions", cleanVer),
	}

	for _, dir := range dirsToCheck {
		if _, err := os.Stat(dir); err == nil {
			iniPath := filepath.Join(dir, "php.ini")
			if _, err := os.Stat(iniPath); err == nil {
				return iniPath
			}
			// If php.ini does not exist, initialize from php.ini-development or production
			devIni := filepath.Join(dir, "php.ini-development")
			prodIni := filepath.Join(dir, "php.ini-production")
			if data, err := os.ReadFile(devIni); err == nil {
				os.WriteFile(iniPath, data, 0644)
				return iniPath
			} else if data, err := os.ReadFile(prodIni); err == nil {
				os.WriteFile(iniPath, data, 0644)
				return iniPath
			}
			return iniPath
		}
	}

	return filepath.Join(baseDir, "php-versions", "php-"+cleanVer, "php.ini")
}

// GetInstalledPHPVersions scans the dist/php-versions directory and returns installed versions
func (a *App) GetInstalledPHPVersions() []PHPVersionInfo {
	var versions []PHPVersionInfo
	phpDir := filepath.Join(a.getDistDir(), "php-versions")

	entries, err := os.ReadDir(phpDir)
	if err != nil {
		return versions
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(phpDir, entry.Name())

		hasPhpExe := false
		if _, err := os.Stat(filepath.Join(fullPath, "php.exe")); err == nil {
			hasPhpExe = true
		}
		hasPhpCgi := false
		if _, err := os.Stat(filepath.Join(fullPath, "php-cgi.exe")); err == nil {
			hasPhpCgi = true
		}

		if hasPhpExe || hasPhpCgi || strings.HasPrefix(entry.Name(), "php-") {
			versionName := strings.TrimPrefix(entry.Name(), "php-")
			status := "Installed"
			if !hasPhpCgi {
				status = "Missing php-cgi.exe"
			} else if !hasPhpExe {
				status = "Missing php.exe"
			}

			versions = append(versions, PHPVersionInfo{
				Version: versionName,
				Path:    fullPath,
				Status:  status,
			})
		}
	}

	return versions
}

// GUI-local structs for Config
type ServerConfig struct {
	APIPort    int    `json:"api_port"`
	DataDir    string `json:"data_dir"`
	PHPWorkers int    `json:"php_workers"`
	APIKey     string `json:"api_key"`
}

type SiteConfig struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	Port         int    `json:"port"`
	DocumentRoot string `json:"document_root"`
	PHPVersion   string `json:"php_version"`
}

type SitesFile struct {
	Sites []SiteConfig `json:"sites"`
}

// GetServerConfig returns the server configuration
func (a *App) GetServerConfig() (ServerConfig, error) {
	var cfg ServerConfig
	configPath := filepath.Join(a.getDistDir(), "config", "server.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		cfg.APIPort = 8081
		cfg.DataDir = a.getDistDir()
		cfg.PHPWorkers = 16
		return cfg, nil
	}
	err = json.Unmarshal(data, &cfg)
	if cfg.PHPWorkers == 0 {
		cfg.PHPWorkers = 16
	}
	// Always reflect actual dist dir
	cfg.DataDir = a.getDistDir()
	return cfg, err
}

// SaveServerConfig saves the server configuration
func (a *App) SaveServerConfig(cfg ServerConfig) error {
	configPath := filepath.Join(a.getDistDir(), "config", "server.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(configPath), 0755)
	return os.WriteFile(configPath, data, 0644)
}

// GetSitesConfig returns the sites configuration
func (a *App) GetSitesConfig() (SitesFile, error) {
	var sf SitesFile
	configPath := filepath.Join(a.getDistDir(), "config", "sites.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		sf.Sites = []SiteConfig{}
		return sf, nil
	}
	err = json.Unmarshal(data, &sf)

	// Auto-heal broken absolute paths for portability (if user moves the folder)
	baseDir := a.getDistDir()
	needsSave := false

	for i, site := range sf.Sites {
		if filepath.IsAbs(site.DocumentRoot) {
			if _, statErr := os.Stat(site.DocumentRoot); os.IsNotExist(statErr) {
				slashPath := filepath.ToSlash(site.DocumentRoot)
				idx := strings.Index(strings.ToLower(slashPath), "/dist/sites/")
				if idx != -1 {
					relPath := slashPath[idx+12:]
					newAbs := filepath.Join(baseDir, "sites", relPath)
					if _, statErr := os.Stat(newAbs); statErr == nil {
						sf.Sites[i].DocumentRoot = newAbs
						needsSave = true
					}
				}
			}
		}
	}

	if needsSave {
		a.SaveSitesConfig(sf)
	}

	return sf, err
}

// SaveSitesConfig saves the sites configuration
func (a *App) SaveSitesConfig(cfg SitesFile) error {
	configPath := filepath.Join(a.getDistDir(), "config", "sites.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(configPath), 0755)
	return os.WriteFile(configPath, data, 0644)
}

// DashboardStats holds the health metrics
type DashboardStats struct {
	Status           string `json:"status"`
	Uptime           string `json:"uptime"`
	ActivePHPWorkers int    `json:"active_php_workers"`
	MemoryAllocated  uint64 `json:"memory_allocated_mb"`
	MemorySys        uint64 `json:"memory_sys_mb"`
	TotalSites       int    `json:"total_sites"`
	APIPort          int    `json:"api_port"`
}

// GetDashboardStats calls the health check API and gets sites count
func (a *App) GetDashboardStats() (DashboardStats, error) {
	var stats DashboardStats

	cfg, _ := a.GetServerConfig()
	stats.APIPort = cfg.APIPort
	if stats.APIPort == 0 {
		stats.APIPort = 8081
	}

	client := http.Client{Timeout: 2 * time.Second}
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = "phforge-secret-key"
	}

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/phforge-health", stats.APIPort)
	req, err := http.NewRequest("GET", healthURL, nil)
	if err == nil {
		req.Header.Set("X-API-Key", apiKey)
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				json.NewDecoder(resp.Body).Decode(&stats)
			} else {
				stats.Status = "Auth Error"
				stats.Uptime = "0s"
			}
		} else {
			stats.Status = "Offline"
			stats.Uptime = "0s"
		}
	} else {
		stats.Status = "Offline"
		stats.Uptime = "0s"
	}

	// Count sites
	sitesCfg, err := a.GetSitesConfig()
	if err == nil {
		stats.TotalSites = len(sitesCfg.Sites)
	}

	return stats, nil
}

// PHPExtension represents a PHP extension status
type PHPExtension struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// GetPHPExtensions parses php.ini and returns the list of extensions
func (a *App) GetPHPExtensions(version string) ([]PHPExtension, error) {
	var extensions []PHPExtension
	iniPath := a.getPHPIniPath(version)

	data, err := os.ReadFile(iniPath)
	if err != nil {
		return extensions, fmt.Errorf("php.ini could not be loaded at %s", iniPath)
	}

	lines := strings.Split(string(data), "\n")
	extMap := make(map[string]bool)

	for _, line := range lines {
		trimLine := strings.TrimSpace(line)
		if trimLine == "" {
			continue
		}

		isCommented := strings.HasPrefix(trimLine, ";")
		content := trimLine
		if isCommented {
			content = strings.TrimSpace(strings.TrimPrefix(trimLine, ";"))
		}

		if strings.HasPrefix(content, "extension=") || strings.HasPrefix(content, "zend_extension=") {
			parts := strings.SplitN(content, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.Trim(name, `"'`)
				name = strings.Split(name, ";")[0]
				name = strings.TrimSpace(name)

				if name == "modulename" || name == "<ext>" || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, ".so") {
					continue
				}

				// OPcache is a Zend Extension managed exclusively via "Tuning & OPcache"
				if strings.EqualFold(name, "opcache") || strings.EqualFold(name, "php_opcache") {
					continue
				}

				extMap[name] = !isCommented
			}
		}
	}

	for name, enabled := range extMap {
		extensions = append(extensions, PHPExtension{
			Name:    name,
			Enabled: enabled,
		})
	}

	return extensions, nil
}

// TogglePHPExtension enables or disables a specific PHP extension in php.ini
func (a *App) TogglePHPExtension(version string, extName string, enable bool) error {
	iniPath := a.getPHPIniPath(version)

	data, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string

	for _, line := range lines {
		trimLine := strings.TrimSpace(line)

		// Always enable extension_dir = "ext" if user is enabling any extension
		if enable && strings.HasPrefix(trimLine, ";") && strings.Contains(trimLine, `extension_dir = "ext"`) {
			line = strings.Replace(line, ";", "", 1)
		}

		// Check for target extension
		isCommented := strings.HasPrefix(trimLine, ";")
		content := trimLine
		if isCommented {
			content = strings.TrimSpace(strings.TrimPrefix(trimLine, ";"))
		}

		if strings.HasPrefix(content, "extension=") || strings.HasPrefix(content, "zend_extension=") {
			parts := strings.SplitN(content, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.Trim(name, `"'`)
				name = strings.Split(name, ";")[0]
				name = strings.TrimSpace(name)

				if name == extName {
					if enable {
						if isCommented {
							idx := strings.Index(line, ";")
							line = line[:idx] + line[idx+1:]
						}
					} else {
						if !isCommented {
							idx := strings.Index(line, content)
							line = line[:idx] + ";" + line[idx:]
						}
					}
				}
			}
		}

		newLines = append(newLines, line)
	}

	output := strings.Join(newLines, "\n")
	return os.WriteFile(iniPath, []byte(output), 0644)
}

// AuditLogEntry represents a parsed log line
type AuditLogEntry struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// GetAuditLogs returns the last 25 lines of the audit log efficiently
func (a *App) GetAuditLogs() []AuditLogEntry {
	var logs []AuditLogEntry
	logPath := filepath.Join(a.getDistDir(), "logs", "audit.log")

	file, err := os.Open(logPath)
	if err != nil {
		return logs
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil || stat.Size() == 0 {
		return logs
	}

	readSize := int64(8192)
	if stat.Size() < readSize {
		readSize = stat.Size()
	}

	buf := make([]byte, readSize)
	_, err = file.ReadAt(buf, stat.Size()-readSize)
	if err != nil {
		return logs
	}

	lines := strings.Split(string(buf), "\n")

	count := 0
	for i := len(lines) - 1; i >= 0 && count < 25; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		entry := AuditLogEntry{
			Time:    "Unknown",
			Type:    "INFO",
			Message: line,
		}

		if len(line) > 20 && line[19] == ' ' {
			entry.Time = line[:19]
			msgPart := line[20:]

			if strings.HasPrefix(msgPart, "[") {
				closeBracket := strings.Index(msgPart, "]")
				if closeBracket != -1 {
					entry.Type = msgPart[1:closeBracket]
					entry.Message = strings.TrimSpace(msgPart[closeBracket+1:])
				}
			} else {
				entry.Message = msgPart
			}
		}

		logs = append(logs, entry)
		count++
	}

	return logs
}

// PHPTuningConfig holds memory, upload, execution time, and OPcache tuning settings
type PHPTuningConfig struct {
	OPcacheEnabled    bool   `json:"opcache_enabled"`
	MemoryLimit       string `json:"memory_limit"`
	UploadMaxFilesize string `json:"upload_max_filesize"`
	PostMaxSize       string `json:"post_max_size"`
	MaxExecutionTime  string `json:"max_execution_time"`
	OPcacheMemory     string `json:"opcache_memory"`
}

// GetPHPTuning parses php.ini and returns the current tuning values
func (a *App) GetPHPTuning(version string) (PHPTuningConfig, error) {
	config := PHPTuningConfig{
		OPcacheEnabled:    false,
		MemoryLimit:       "128M",
		UploadMaxFilesize: "2M",
		PostMaxSize:       "8M",
		MaxExecutionTime:  "30",
		OPcacheMemory:     "128",
	}

	iniPath := a.getPHPIniPath(version)
	data, err := os.ReadFile(iniPath)
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(data), "\n")
	hasZendOpcache := false
	opcacheEnabledDirectly := false

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		isCommented := strings.HasPrefix(line, ";")
		content := line
		if isCommented {
			content = strings.TrimSpace(strings.TrimPrefix(line, ";"))
		}

		parts := strings.SplitN(content, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(strings.Split(parts[1], ";")[0])
			v = strings.Trim(v, `"'`)

			switch k {
			case "zend_extension":
				if !isCommented && (v == "opcache" || strings.Contains(v, "php_opcache")) {
					hasZendOpcache = true
				}
			case "opcache.enable":
				if !isCommented && (v == "1" || strings.EqualFold(v, "on")) {
					opcacheEnabledDirectly = true
				}
			case "memory_limit":
				if !isCommented {
					config.MemoryLimit = v
				}
			case "upload_max_filesize":
				if !isCommented {
					config.UploadMaxFilesize = v
				}
			case "post_max_size":
				if !isCommented {
					config.PostMaxSize = v
				}
			case "max_execution_time":
				if !isCommented {
					config.MaxExecutionTime = v
				}
			case "opcache.memory_consumption":
				if !isCommented {
					config.OPcacheMemory = v
				}
			}
		}
	}

	config.OPcacheEnabled = hasZendOpcache && opcacheEnabledDirectly
	return config, nil
}

// SavePHPTuning writes OPcache and resource limits to php.ini
func (a *App) SavePHPTuning(version string, tuning PHPTuningConfig) error {
	iniPath := a.getPHPIniPath(version)
	data, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string

	keysHandled := map[string]bool{
		"memory_limit":                    false,
		"upload_max_filesize":             false,
		"post_max_size":                   false,
		"max_execution_time":              false,
		"zend_extension":                  false,
		"opcache.enable":                  false,
		"opcache.enable_cli":              false,
		"opcache.memory_consumption":      false,
		"opcache.interned_strings_buffer": false,
		"opcache.max_accelerated_files":   false,
		"opcache.validate_timestamps":     false,
	}

	opcacheMemory := tuning.OPcacheMemory
	if opcacheMemory == "" {
		opcacheMemory = "128"
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		isCommented := strings.HasPrefix(line, ";")
		content := line
		if isCommented {
			content = strings.TrimSpace(strings.TrimPrefix(line, ";"))
		}

		parts := strings.SplitN(content, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(strings.Split(parts[1], ";")[0])
			v = strings.Trim(v, `"'`)

			switch k {
			case "memory_limit":
				newLines = append(newLines, fmt.Sprintf("memory_limit = %s", tuning.MemoryLimit))
				keysHandled["memory_limit"] = true
				continue
			case "upload_max_filesize":
				newLines = append(newLines, fmt.Sprintf("upload_max_filesize = %s", tuning.UploadMaxFilesize))
				keysHandled["upload_max_filesize"] = true
				continue
			case "post_max_size":
				newLines = append(newLines, fmt.Sprintf("post_max_size = %s", tuning.PostMaxSize))
				keysHandled["post_max_size"] = true
				continue
			case "max_execution_time":
				newLines = append(newLines, fmt.Sprintf("max_execution_time = %s", tuning.MaxExecutionTime))
				keysHandled["max_execution_time"] = true
				continue
			case "zend_extension":
				if v == "opcache" || strings.Contains(v, "php_opcache") {
					if tuning.OPcacheEnabled {
						newLines = append(newLines, "zend_extension=opcache")
					} else {
						newLines = append(newLines, ";zend_extension=opcache")
					}
					keysHandled["zend_extension"] = true
					continue
				}
			case "opcache.enable":
				if tuning.OPcacheEnabled {
					newLines = append(newLines, "opcache.enable=1")
				} else {
					newLines = append(newLines, ";opcache.enable=0")
				}
				keysHandled["opcache.enable"] = true
				continue
			case "opcache.enable_cli":
				if tuning.OPcacheEnabled {
					newLines = append(newLines, "opcache.enable_cli=0")
				} else {
					newLines = append(newLines, ";opcache.enable_cli=0")
				}
				keysHandled["opcache.enable_cli"] = true
				continue
			case "opcache.memory_consumption":
				if tuning.OPcacheEnabled {
					newLines = append(newLines, fmt.Sprintf("opcache.memory_consumption=%s", opcacheMemory))
				} else {
					newLines = append(newLines, fmt.Sprintf(";opcache.memory_consumption=%s", opcacheMemory))
				}
				keysHandled["opcache.memory_consumption"] = true
				continue
			case "opcache.interned_strings_buffer":
				if tuning.OPcacheEnabled {
					newLines = append(newLines, "opcache.interned_strings_buffer=8")
				} else {
					newLines = append(newLines, ";opcache.interned_strings_buffer=8")
				}
				keysHandled["opcache.interned_strings_buffer"] = true
				continue
			case "opcache.max_accelerated_files":
				if tuning.OPcacheEnabled {
					newLines = append(newLines, "opcache.max_accelerated_files=10000")
				} else {
					newLines = append(newLines, ";opcache.max_accelerated_files=10000")
				}
				keysHandled["opcache.max_accelerated_files"] = true
				continue
			case "opcache.validate_timestamps":
				if tuning.OPcacheEnabled {
					newLines = append(newLines, "opcache.validate_timestamps=1")
				} else {
					newLines = append(newLines, ";opcache.validate_timestamps=1")
				}
				keysHandled["opcache.validate_timestamps"] = true
				continue
			}
		}

		newLines = append(newLines, rawLine)
	}

	var unhandled []string
	if !keysHandled["memory_limit"] {
		unhandled = append(unhandled, fmt.Sprintf("memory_limit = %s", tuning.MemoryLimit))
	}
	if !keysHandled["upload_max_filesize"] {
		unhandled = append(unhandled, fmt.Sprintf("upload_max_filesize = %s", tuning.UploadMaxFilesize))
	}
	if !keysHandled["post_max_size"] {
		unhandled = append(unhandled, fmt.Sprintf("post_max_size = %s", tuning.PostMaxSize))
	}
	if !keysHandled["max_execution_time"] {
		unhandled = append(unhandled, fmt.Sprintf("max_execution_time = %s", tuning.MaxExecutionTime))
	}

	if tuning.OPcacheEnabled {
		if !keysHandled["zend_extension"] {
			unhandled = append(unhandled, "zend_extension=opcache")
		}
		if !keysHandled["opcache.enable"] {
			unhandled = append(unhandled, "opcache.enable=1")
		}
		if !keysHandled["opcache.memory_consumption"] {
			unhandled = append(unhandled, fmt.Sprintf("opcache.memory_consumption=%s", opcacheMemory))
		}
		if !keysHandled["opcache.interned_strings_buffer"] {
			unhandled = append(unhandled, "opcache.interned_strings_buffer=8")
		}
		if !keysHandled["opcache.max_accelerated_files"] {
			unhandled = append(unhandled, "opcache.max_accelerated_files=10000")
		}
		if !keysHandled["opcache.validate_timestamps"] {
			unhandled = append(unhandled, "opcache.validate_timestamps=1")
		}
	}

	if len(unhandled) > 0 {
		newLines = append(newLines, "", "; --- PHPForge Performance Optimization ---")
		newLines = append(newLines, unhandled...)
	}

	output := strings.Join(newLines, "\n")
	return os.WriteFile(iniPath, []byte(output), 0644)
}
