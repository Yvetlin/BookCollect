package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// 1) Берём конфиг из окружения
	// Приоритет: DATABASE_URL > POSTGRES_DSN > сборка из отдельных переменных
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		host := getenv("POSTGRES_HOST", "127.0.0.1")
		port := getenv("POSTGRES_PORT", "5432")
		user := getenv("POSTGRES_USER", "postgres")
		pass := os.Getenv("POSTGRES_PASSWORD")
		name := getenv("POSTGRES_DB", "BookCollect")
		sslm := getenv("POSTGRES_SSLMODE", "disable") // локально disable; в проде обычно require/verify-full

		// lib/pq key=value формат:
		// ВАЖНО: не печатай пароль в логах
		parts := []string{
			"host=" + host,
			"port=" + port,
			"user=" + user,
			"dbname=" + name,
			"sslmode=" + sslm,
		}
		if pass != "" {
			parts = append(parts, "password="+pass)
		}
		dsn = strings.Join(parts, " ")
	}

	// 2) Подключаемся
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db: open failed: %v", err)
	}

	// 3) Пул коннектов
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	// 4) Ping с таймаутом (не вешаем процесс)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := DB.PingContext(ctx); err != nil {
		log.Fatalf("db: ping failed: %v", err)
	}

	// 5) Логируем безопасно (без пароля/полного DSN)
	logSafeDSN()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func logSafeDSN() {
	// Печатаем только «куда», без секретов
	host := getenv("POSTGRES_HOST", "")
	user := getenv("POSTGRES_USER", "")
	dbn := getenv("POSTGRES_DB", "")
	if host == "" && user == "" && dbn == "" {
		// возможно пришёл DATABASE_URL; выдернем хост/базу по-минимуму
		u := os.Getenv("DATABASE_URL")
		if u != "" {
			// не разбираем полноценно, просто не палим пароль
			log.Println("db: connected (DATABASE_URL provided)")
			return
		}
	}
	log.Printf("db: connected (host=%s user=%s db=%s)", host, user, dbn)
}
