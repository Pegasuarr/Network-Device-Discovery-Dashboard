package configs

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	RedisHost      string
	RedisPort      string
	RedisPassword  string
	JWTSecret      string
	JWTAccessTTL   int // in minutes
	JWTRefreshTTL  int // in days
	RateLimitLimit int // limit per minute
}

func LoadConfig() *Config {
	loadEnvFile(".env")

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "network_monitor"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		RedisHost:      getEnv("REDIS_HOST", "localhost"),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:      getEnv("JWT_SECRET", "super_secret_jwt_key_for_network_dashboard"),
		JWTAccessTTL:   getEnvAsInt("JWT_ACCESS_TTL", 60),      // 60 minutes
		JWTRefreshTTL:  getEnvAsInt("JWT_REFRESH_TTL", 7),      // 7 days
		RateLimitLimit: getEnvAsInt("RATE_LIMIT_LIMIT", 100),   // 100 requests per minute
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // Silently skip if no .env file exists
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Strip surrounding quotes
			val = strings.Trim(val, `"'`)
			os.Setenv(key, val)
		}
	}
}
