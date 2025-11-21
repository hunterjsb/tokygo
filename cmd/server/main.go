package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hunterjsb/tokygo/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get the root directory (two levels up from cmd/server)
	rootDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	// Create server and register handlers
	server := internal.NewServer(rootDir)
	server.RegisterHandlers()

	addr := fmt.Sprintf(":%s", port)
	frontendDir := filepath.Join(rootDir, "frontend")

	fmt.Printf("🚀 Server starting on http://localhost%s\n", addr)
	fmt.Printf("📂 Serving frontend from: %s\n", frontendDir)
	fmt.Printf("🗾 View the map at: http://localhost%s/\n", addr)
	fmt.Println("\nPress Ctrl+C to stop the server")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
