package main

import (
	"log"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/anyzearch/admin/core/internal/server"
)

func main() {
	global.InitConfig()
	global.InitDB()
	r := server.NewRouter()

	address := "0.0.0.0:" + global.ServerPort
	log.Printf("Anyzearch Admin Server starting on %s\n", address)
	if err := r.Run(address); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
