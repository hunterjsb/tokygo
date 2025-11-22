package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hunterjsb/tokygo/internal"
)

// loadEnv loads environment variables from .env file
func loadEnv() error {
	file, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func main() {
	// Load .env file
	if err := loadEnv(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get the root directory (find it by looking for go.mod)
	rootDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	// If we're in cmd/server, go up two levels
	if filepath.Base(rootDir) == "server" {
		rootDir = filepath.Join(rootDir, "../..")
		rootDir, err = filepath.Abs(rootDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create server and register handlers
	server := internal.NewServer(rootDir)
	server.RegisterHandlers()

	addr := fmt.Sprintf(":%s", port)
	frontendDir := filepath.Join(rootDir, "frontend", "dist")

	fmt.Printf("🚀 Server starting on http://localhost%s\n", addr)
	fmt.Printf("📂 Serving frontend from: %s\n", frontendDir)
	fmt.Printf("🗾 View the map at: http://localhost%s/\n", addr)
	fmt.Println("\nPress Ctrl+C to stop the server")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
