package config

import (
"bufio"
"os"
"strings"
)

type Config struct {
Port        string
DatabaseURL string
JWTSecret   string
Environment string
CorsOrigins string
StoragePath string
}

func loadEnvFile(path string) {
file, err := os.Open(path)
if err != nil {
return
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
line := strings.TrimSpace(scanner.Text())
if line == "" || strings.HasPrefix(line, "#") {
continue
}
parts := strings.SplitN(line, "=", 2)
if len(parts) == 2 {
key := strings.TrimSpace(parts[0])
val := strings.TrimSpace(parts[1])
val = strings.Trim(val, `"'`)
if os.Getenv(key) == "" {
os.Setenv(key, val)
}
}
}
}

func LoadConfig() *Config {
loadEnvFile(".env")
loadEnvFile("/var/www/salenmobilebackend/.env")

port := os.Getenv("PORT")
if port == "" {
port = "8085"
}

dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
dbURL = "salenmobile:SalenMobile@DB2026!@tcp(127.0.0.1:3306)/salenmobile_db?charset=utf8mb4&parseTime=True&loc=Local"
}

jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
jwtSecret = "salen-mobile-guru-secret-key-2026"
}

env := os.Getenv("APP_ENV")
if env == "" {
env = "production"
}

corsOrigins := os.Getenv("CORS_ORIGINS")
if corsOrigins == "" {
corsOrigins = "*"
}

storagePath := os.Getenv("STORAGE_PATH")
if storagePath == "" {
storagePath = "/var/www/salenmobilebackend/storage"
}

return &Config{
Port:        port,
DatabaseURL: dbURL,
JWTSecret:   jwtSecret,
Environment: env,
CorsOrigins: corsOrigins,
StoragePath: storagePath,
}
}
