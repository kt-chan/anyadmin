package main

import (
	"log"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/server"
	"anyadmin-backend/pkg/utils"
)

func main() {
	utils.InitLogger()
	global.InitConfig()
	// Initialize Data Store instead of real DB
	utils.InitData()

	r := server.NewRouter()

	address := "0.0.0.0:" + global.ServerPort
	log.Printf("AnythingLLM Admin Backend (Mock Mode) starting on %s\n", address)
	if err := r.Run(address); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
