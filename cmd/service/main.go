package main

import (
	"database/sql"
	"fmt"
	"log"
	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/handlers"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/stan.go"
	_ "github.com/lib/pq"
)

type App struct {
	config   *config.Config
	db       *sql.DB
	cache    *cache.Cache
	service  service.OrderService
	handlers *handlers.Handler
	stanConn stan.Conn
}

func main() {
	cfg := config.Load()
	
	app := &App{
		config: cfg,
		cache:  cache.New(),
	}

	if err := app.initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer app.db.Close()

	if err := app.restoreCache(); err != nil {
		log.Fatal("Failed to restore cache:", err)
	}

	if err := app.initNATS(); err != nil {
		log.Fatal("Failed to initialize NATS:", err)
	}
	defer app.stanConn.Close()

	app.service = service.New(app.db, app.cache, app.stanConn)
	app.handlers = handlers.New(app.service)
	
	app.startHTTPServer()
}

func (app *App) initDB() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		app.config.DBHost, app.config.DBPort, app.config.DBUser, 
		app.config.DBPassword, app.config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	app.db = db
	log.Println("Connected to database successfully")
	return nil
}

func (app *App) initNATS() error {
	sc, err := stan.Connect(app.config.NATSClusterID, app.config.NATSClientID)
	if err != nil {
		return err
	}

	app.stanConn = sc

	_, err = sc.Subscribe(app.config.NATSChannel, func(msg *stan.Msg) {
		if err := app.service.ProcessMessage(msg.Data); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}, stan.DurableName(app.config.NATSDurableID), stan.SetManualAckMode())

	if err != nil {
		return err
	}

	log.Println("Subscribed to NATS channel successfully")
	return nil
}

func (app *App) restoreCache() error {
	tempService := service.New(app.db, app.cache, nil)
	return tempService.RestoreCache()
}

func (app *App) startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	router.GET("/api/order/:id", app.handlers.GetOrder)
	router.GET("/api/health", app.handlers.HealthCheck)

	router.GET("/", app.handlers.WebInterface)

	log.Printf("HTTP server starting on :%s", app.config.HTTPPort)
	log.Fatal(router.Run(":" + app.config.HTTPPort))
}