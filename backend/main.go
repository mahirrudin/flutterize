package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/flutterize/backend/config"
	"github.com/flutterize/backend/handler"
	"github.com/flutterize/backend/middleware"
	"github.com/flutterize/backend/repository"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	// Validate required environment variables
	if cfg.DBPass == "" || cfg.DBRootPass == "" || cfg.JWTSecret == "" {
		log.Fatal("Required environment variables not set: DB_PASS, MYSQL_ROOT_PASSWORD, JWT_SECRET")
	}

	// Connect to database, run migrations/seed as root, then reconnect as API user
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	pointsRepo := repository.NewPointsRepository(db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, cfg)
	userHandler := handler.NewUserHandler(userRepo)
	pointsHandler := handler.NewPointsHandler(pointsRepo, userRepo)

	// Setup routes
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/api/reset-password", authHandler.ResetPassword)

	// Protected routes
	mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.AuthMiddleware(cfg.JWTSecret, userHandler.GetProfile)(w, r)
		case http.MethodPut:
			middleware.AuthMiddleware(cfg.JWTSecret, userHandler.UpdateProfile)(w, r)
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/change-password",
		middleware.AuthMiddleware(cfg.JWTSecret, userHandler.ChangePassword))

	mux.HandleFunc("/api/points/balance",
		middleware.AuthMiddleware(cfg.JWTSecret, pointsHandler.GetBalance))

	mux.HandleFunc("/api/points/transfer",
		middleware.AuthMiddleware(cfg.JWTSecret, pointsHandler.TransferPoints))

	mux.HandleFunc("/api/points/history",
		middleware.AuthMiddleware(cfg.JWTSecret, pointsHandler.GetTransactionHistory))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply global middleware
	var finalHandler http.Handler = mux
	finalHandler = middleware.CORSMiddleware(finalHandler)
	finalHandler = middleware.LoggingMiddleware(finalHandler)

	addr := ":" + cfg.ServerPort
	log.Printf("Starting Flutterize API server on %s", addr)

	// Check if TLS cert files actually exist on disk
	tlsEnabled := false
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		_, certErr := os.Stat(cfg.TLSCert)
		_, keyErr := os.Stat(cfg.TLSKey)
		tlsEnabled = certErr == nil && keyErr == nil
	}

	if tlsEnabled {
		log.Printf("TLS enabled with cert=%s key=%s", cfg.TLSCert, cfg.TLSKey)
		if err := http.ListenAndServeTLS(addr, cfg.TLSCert, cfg.TLSKey, finalHandler); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		log.Println("WARNING: TLS certs not found, running in plain HTTP mode")
		if err := http.ListenAndServe(addr, finalHandler); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	rootDSN := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		cfg.DBRootPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var db *sql.DB
	var err error

	// Step 1: Connect as root with retry
	for i := 0; i < 30; i++ {
		db, err = sql.Open("mysql", rootDSN)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Connected to database as root")
				break
			}
		}
		log.Printf("Waiting for database... attempt %d/30: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect after 30 attempts: %w", err)
	}

	// Step 2: Run schema migrations AS ROOT (needs CREATE privilege)
	runMigrations(db)

	// Step 3: Create dedicated API user AS ROOT
	createAPIUser(db, cfg)

	// Step 4: Seed default users AS ROOT
	seedUsers(db)

	// Step 5: Reconnect with the least-privilege API user
	db.Close()
	apiDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err = sql.Open("mysql", apiDSN)
	if err == nil {
		err = db.Ping()
		if err == nil {
			log.Println("Reconnected with API user successfully")
			return db, nil
		}
	}

	// Fallback to root if API user fails
	log.Printf("Failed to connect with API user, falling back to root: %v", err)
	db, err = sql.Open("mysql", rootDSN)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func createAPIUser(db *sql.DB, cfg *config.Config) {
	migrationFile := "migration/002_create_api_user.sql"
	data, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Printf("Could not read API user migration: %v", err)
		return
	}

	// Replace placeholder with actual password from environment variable
	content := strings.ReplaceAll(string(data), "${DB_PASS}", cfg.DBPass)
	_, err = db.Exec(content)
	if err != nil {
		log.Printf("API user migration warning: %v", err)
	} else {
		log.Println("API user created/verified successfully")
	}
}

func runMigrations(db *sql.DB) {
	data, err := os.ReadFile("migration/001_schema.sql")
	if err != nil {
		log.Printf("Could not read schema migration: %v", err)
		return
	}

	_, err = db.Exec(string(data))
	if err != nil {
		log.Printf("Schema migration warning: %v", err)
	} else {
		log.Println("Schema migration completed successfully")
	}
}

func seedUsers(db *sql.DB) {
	seeds := []struct {
		email     string
		phone     string
		fullname  string
		birthdate string
		password  string
		points    int64
	}{
		{
			os.Getenv("SEED_ADMIN_EMAIL"),
			os.Getenv("SEED_ADMIN_PHONE"),
			"Admin User",
			"1990-01-01",
			os.Getenv("SEED_ADMIN_PASSWORD"),
			getEnvInt("SEED_ADMIN_POINTS", 1000),
		},
		{
			os.Getenv("SEED_USER_EMAIL"),
			os.Getenv("SEED_USER_PHONE"),
			"Standard User",
			"1995-01-01",
			os.Getenv("SEED_USER_PASSWORD"),
			getEnvInt("SEED_USER_POINTS", 500),
		},
	}

	for _, s := range seeds {
		if s.email == "" || s.password == "" {
			continue
		}

		var count int
		db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", s.email).Scan(&count)
		if count > 0 {
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", s.email, err)
			continue
		}

		_, err = db.Exec(
			"INSERT INTO users (email, phone, fullname, birthdate, password, points_balance) VALUES (?, ?, ?, ?, ?, ?)",
			s.email, s.phone, s.fullname, s.birthdate, string(hashedPassword), s.points,
		)
		if err != nil {
			log.Printf("Failed to seed user %s: %v", s.email, err)
		} else {
			log.Printf("Seeded user: %s (points: %d)", s.email, s.points)
		}
	}
}

func getEnvInt(key string, fallback int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}
