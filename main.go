package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"olivermdelgado/url-shortener/pkg/controller"
	"olivermdelgado/url-shortener/pkg/model"
)

func main() {
	// initialize with a redis client
	cfg := controller.Config{
		RDS: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		Address: "http://localhost:8080",
	}
	// attempt connection to postgress
	psqlURI := "postgresql://localhost/shorty"
	db, err := gorm.Open(postgres.Open(psqlURI), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	cfg.DB= db

	// this is how you need to close db connections in gorm v2
	sqlDB, err := db.DB()
	defer sqlDB.Close()

	// set up postgresql db to include any new tables that may be needed on start up
	if err = model.Migrate(db); err != nil {
		log.Fatal(err.Error())
	}

	// stand up routes and start server
	e := echo.New()
	e.Use(middleware.Logger())

	e.POST("/", cfg.CreateNewKeyValue)
	e.GET("/:short_key", cfg.GetKeyValue)
	// TODO: e.GET("/:short_key/metrics", cfg.getURLMetrics)

	e.Logger.Fatal(e.Start(":8080"))
}