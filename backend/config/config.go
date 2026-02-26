package config

import "os"

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPass        string
	DBName        string
	DBRootPass    string
	JWTSecret     string
	TLSCert       string
	TLSKey        string
	SMTPHost      string
	SMTPPort      string
	SMSGatewayURL string
	ServerPort    string
}

func Load() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", ""),
		DBPass:        getEnv("DB_PASS", ""),
		DBName:        getEnv("DB_NAME", "flutterize"),
		DBRootPass:    getEnv("MYSQL_ROOT_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		TLSCert:       getEnv("TLS_CERT", ""),
		TLSKey:        getEnv("TLS_KEY", ""),
		SMTPHost:      getEnv("SMTP_HOST", "localhost"),
		SMTPPort:      getEnv("SMTP_PORT", "2525"),
		SMSGatewayURL: getEnv("SMS_GATEWAY_URL", "http://localhost:7600"),
		ServerPort:    getEnv("SERVER_PORT", "8443"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
