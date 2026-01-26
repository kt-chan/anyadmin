package service

import (
	"log"
	"time"
)

func ControlContainer(containerName string, action string) error {
	log.Printf("[Container] Mock action: %s on container: %s", action, containerName)
	time.Sleep(500 * time.Millisecond)
	return nil
}
