package main

import (
	"log"

	"anyadmin-backend/internal/global"
	"anyadmin-backend/internal/mockdata"
	"anyadmin-backend/internal/server"
)

func main() {
	global.InitConfig()
	// Initialize Mock Data Store instead of real DB
	mockdata.InitData()

	r := server.NewRouter()

	address := "0.0.0.0:" + global.ServerPort
	log.Printf("AnythingLLM Admin Backend (Mock Mode) starting on %s\n", address)
	if err := r.Run(address); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
