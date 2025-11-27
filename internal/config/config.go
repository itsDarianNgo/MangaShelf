package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the full application configuration loaded from file, environment, and flags.
type Config struct {
	Server        ServerConfig       `mapstructure:"server"`
	Library       LibraryConfig      `mapstructure:"library"`
	Downloader    DownloaderConfig   `mapstructure:"downloader"`
	Formats       FormatConfig       `mapstructure:"formats"`
	Updates       UpdateConfig       `mapstructure:"updates"`
	Metadata      MetadataConfig     `mapstructure:"metadata"`
	Notifications NotificationConfig `mapstructure:"notifications"`
	Reader        ReaderConfig       `mapstructure:"reader"`
	Sources       SourceConfig       `mapstructure:"sources"`
	Logging       LoggingConfig      `mapstructure:"logging"`
	Database      DatabaseConfig     `mapstructure:"database"`
}

type ServerConfig struct {
	Host    string     `mapstructure:"host"`
	Port    int        `mapstructure:"port"`
	BaseURL string     `mapstructure:"baseUrl"`
	CORS    CORSConfig `mapstructure:"cors"`
}

type CORSConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Origins []string `mapstructure:"origins"`
}

type LibraryConfig struct {
	Path            string `mapstructure:"path"`
	ScanOnStartup   bool   `mapstructure:"scanOnStartup"`
	WatchForChanges bool   `mapstructure:"watchForChanges"`
}

type DownloaderConfig struct {
	Workers       int           `mapstructure:"workers"`
	RetryAttempts int           `mapstructure:"retryAttempts"`
	RetryDelay    time.Duration `mapstructure:"retryDelay"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RateLimit     string        `mapstructure:"rateLimit"`
	UserAgent     string        `mapstructure:"userAgent"`
}

type FormatConfig struct {
	Default           string `mapstructure:"default"`
	CompressImages    bool   `mapstructure:"compressImages"`
	JPEGQuality       int    `mapstructure:"jpegQuality"`
	MaxImageWidth     int    `mapstructure:"maxImageWidth"`
	GenerateComicInfo bool   `mapstructure:"generateComicInfo"`
}

type UpdateConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	DefaultInterval string `mapstructure:"defaultInterval"`
	CheckOnStartup  bool   `mapstructure:"checkOnStartup"`
	AutoDownload    bool   `mapstructure:"autoDownload"`
}

type MetadataConfig struct {
	FetchAnilist   bool          `mapstructure:"fetchAnilist"`
	DownloadCovers bool          `mapstructure:"downloadCovers"`
	CoverSize      string        `mapstructure:"coverSize"`
	Anilist        AnilistConfig `mapstructure:"anilist"`
}

type AnilistConfig struct {
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
}

type NotificationConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Apprise AppriseConfig `mapstructure:"apprise"`
	Events  EventConfig   `mapstructure:"events"`
}

type AppriseConfig struct {
	URLs []string `mapstructure:"urls"`
}

type EventConfig struct {
	NewChapters      bool `mapstructure:"newChapters"`
	DownloadComplete bool `mapstructure:"downloadComplete"`
	DownloadFailed   bool `mapstructure:"downloadFailed"`
}

type ReaderConfig struct {
	DefaultMode      string `mapstructure:"defaultMode"`
	DefaultDirection string `mapstructure:"defaultDirection"`
	PreloadPages     int    `mapstructure:"preloadPages"`
	SaveProgress     bool   `mapstructure:"saveProgress"`
}

type SourceConfig struct {
	CustomPath string         `mapstructure:"customPath"`
	Default    string         `mapstructure:"default"`
	Mangadex   MangadexConfig `mapstructure:"mangadex"`
}

type MangadexConfig struct {
	Language        string `mapstructure:"language"`
	NSFW            bool   `mapstructure:"nsfw"`
	ShowUnavailable bool   `mapstructure:"showUnavailable"`
}

type LoggingConfig struct {
	Level  string            `mapstructure:"level"`
	Format string            `mapstructure:"format"`
	File   LoggingFileConfig `mapstructure:"file"`
}

type LoggingFileConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Path       string `mapstructure:"path"`
	MaxSize    string `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
}

type DatabaseConfig struct {
	Path    string `mapstructure:"path"`
	WALMode bool   `mapstructure:"walMode"`
}

// Load reads configuration from file, environment variables, and defaults.
func Load(configPath string) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetConfigType("yaml")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		for _, dir := range defaultConfigPaths() {
			v.AddConfigPath(dir)
		}
		v.SetConfigName("config")
	}

	v.SetEnvPrefix("mangashelf")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func defaultConfigPaths() []string {
	paths := []string{".", filepath.Join(".", "config"), "/data"}
	if cfgDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(cfgDir, "mangashelf"))
	}
	return paths
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.baseUrl", "")
	v.SetDefault("server.cors.enabled", false)
	v.SetDefault("server.cors.origins", []string{})

	v.SetDefault("library.path", "./data/manga")
	v.SetDefault("library.scanOnStartup", true)
	v.SetDefault("library.watchForChanges", false)

	v.SetDefault("downloader.workers", 3)
	v.SetDefault("downloader.retryAttempts", 3)
	v.SetDefault("downloader.retryDelay", "5s")
	v.SetDefault("downloader.timeout", "30s")
	v.SetDefault("downloader.rateLimit", "2/s")
	v.SetDefault("downloader.userAgent", "MangaShelf/1.0")

	v.SetDefault("formats.default", "cbz")
	v.SetDefault("formats.compressImages", false)
	v.SetDefault("formats.jpegQuality", 85)
	v.SetDefault("formats.maxImageWidth", 0)
	v.SetDefault("formats.generateComicInfo", true)

	v.SetDefault("updates.enabled", true)
	v.SetDefault("updates.defaultInterval", "0 */6 * * *")
	v.SetDefault("updates.checkOnStartup", true)
	v.SetDefault("updates.autoDownload", true)

	v.SetDefault("metadata.fetchAnilist", true)
	v.SetDefault("metadata.downloadCovers", true)
	v.SetDefault("metadata.coverSize", "large")
	v.SetDefault("metadata.anilist.clientId", "")
	v.SetDefault("metadata.anilist.clientSecret", "")

	v.SetDefault("notifications.enabled", false)
	v.SetDefault("notifications.apprise.urls", []string{})
	v.SetDefault("notifications.events.newChapters", true)
	v.SetDefault("notifications.events.downloadComplete", false)
	v.SetDefault("notifications.events.downloadFailed", true)

	v.SetDefault("reader.defaultMode", "single")
	v.SetDefault("reader.defaultDirection", "rtl")
	v.SetDefault("reader.preloadPages", 2)
	v.SetDefault("reader.saveProgress", true)

	v.SetDefault("sources.customPath", "./data/scrapers")
	v.SetDefault("sources.default", "mangadex")
	v.SetDefault("sources.mangadex.language", "en")
	v.SetDefault("sources.mangadex.nsfw", false)
	v.SetDefault("sources.mangadex.showUnavailable", false)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
	v.SetDefault("logging.file.enabled", false)
	v.SetDefault("logging.file.path", "./data/logs/mangashelf.log")
	v.SetDefault("logging.file.maxSize", "100MB")
	v.SetDefault("logging.file.maxBackups", 3)

	v.SetDefault("database.path", "./data/mangashelf.db")
	v.SetDefault("database.walMode", true)
}
