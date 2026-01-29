package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// InitLogger configures the standard logger to write to both stdout and a file.
func InitLogger() {
	// Define log directory and file
	// Assuming running from backend/ directory or root where backend is a subdir.
	// We'll try to find the project root or just use a relative path.
	// Using "../logs" assumes we are running from "backend/" directory.
	logDir := "../logs"
	logFile := "backend.log"

	// Ensure logs directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return
	}

	// Open log file
	file, err := os.OpenFile(filepath.Join(logDir, logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	// Set log output to multi-writer (stdout + file)
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Logger initialized. Writing to stdout and", filepath.Join(logDir, logFile))
}
