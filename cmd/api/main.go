package main

import (
	"context"
	"database/sql"
	"flag"
	"runtime"

	"fmt"
	_ "github.com/lib/pq"
	// "log"
	"movies.cosmasgithinji.net/internal/data"
	"movies.cosmasgithinji.net/internal/jsonlog"
	"movies.cosmasgithinji.net/internal/mailer"

	// "net/http"
	"expvar"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"sync"
	"time"
)

// const version = "1.0.0"

var (
	buildTime string
	version string
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	// logger *log.Logger
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	err := godotenv.Load(".env")
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", getEnvAsInt("api_port", 8000), "API Server port")
	flag.StringVar(&cfg.env, "env", getEnvAsString("environment", "production"), "Environment (developement|staging|production)")

	// flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://movies:moviespass@localhost/movies?sslmode=disable", "PostgreSQL DSN")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("db_dsn"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", getEnvAsInt("db_max_open_conns", 25), "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", getEnvAsInt("db_max_idle_conns", 25), "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", getEnvAsString("db_max_idle_time", "15m"), "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", getEnvAsFloat("limiter_rps", 2), "Rate limiter maximu requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", getEnvAsInt("limiter_burst", 4), "rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", getEnvAsBool("limiter_enabled", true), "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", getEnvAsString("smtp_host", "smtp.mailtrap.io"), "SMTP Host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", getEnvAsInt("smtp_port", 25), "SMTP Port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("smtp_username"), "SMTP Username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("smtp_password"), "SMTP Password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("smtp_sender"), "SMTP Sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build Time:\t%s\n", buildTime)
		os.Exit(0)
	}

	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// srv := &http.Server{
	// 	Addr:         fmt.Sprintf(":%d", cfg.port),
	// 	Handler:      app.routes(),
	// 	ErrorLog:     log.New(logger, "", 0),
	// 	IdleTimeout:  time.Minute,
	// 	ReadTimeout:  10 * time.Second,
	// 	WriteTimeout: 30 * time.Second,
	// }

	// logger.PrintInfo("starting server", map[string]string{
	// 	"addr": srv.Addr,
	// 	"env":  cfg.env,
	// })
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
