package main

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
	"github.com/raihaninkam/tickitz/internals/configs"
	"github.com/raihaninkam/tickitz/internals/routers"
	"github.com/raihaninkam/tickitz/pkg"
)

// @title 					Backend Golang Tickitz App
// @version 				1.0
// @host						localhost:9001
// @basePath				/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan format: Bearer <token>
func main() {

	// Init Database
	db, err := configs.InitDB()
	if err != nil {
		log.Println("Failed to connect to database\nCause: ", err.Error())
	}
	defer db.Close()

	if err := configs.PingDB(db); err != nil {
		log.Println("Ping to DB failed\nCause: ", err.Error())
		return
	}
	log.Println("db connected")

	// Init Redis
	rdb, err := configs.InitRedis()
	if err != nil {
		log.Println("Ping to Redis failed\nCause: ", err.Error())
		return
	}

	hc := pkg.NewHashConfig()
	hc.UseRecommended()

	router := routers.InitRouter(db, rdb, hc)

	router.Static("/public", "./public")

	router.Run("localhost:9001")
}
