package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/driver"
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func buildDSN() string {
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "5432")
	user := env("DB_USER", "app")
	name := env("DB_NAME", "appdb")
	ssl := env("DB_SSLMODE", "disable")

	parts := []string{
		"host=" + host,
		"port=" + port,
		"user=" + user,
		"dbname=" + name,
		"sslmode=" + ssl,
	}

	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		parts = append(parts, "password="+pass)
	}

	if extra := os.Getenv("DB_EXTRA"); extra != "" {
		parts = append(parts, extra)
	}

	return strings.Join(parts, " ")
}

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()

	defer close(app.MailChan)

	fmt.Println("Starting mail listener...")
	listenForMail()

	addr := ":" + env("PORT", "8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: routes(&app),
	}

	infoLog.Printf("HTTP server listening on %s (env=%s)\n", addr, env("APP_ENV", "dev"))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errorLog.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = env("APP_ENV", "dev") == "prod"

	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()

	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	infoLog.Println("Connecting to database...")
	dsn := buildDSN()
	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %s", err)
	}

	infoLog.Println("Connected to database")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("cannot create template cache: %s", err)
	}
	app.TemplateCache = tc

	app.UseCache = env("USE_TEMPLATE_CACHE", "false") == "true"

	repo := handlers.NewRepo(&app, db)

	handlers.NewHandlers(repo)

	render.NewRenderer(&app)

	helpers.NewHelpers(&app)

	return db, nil
}
