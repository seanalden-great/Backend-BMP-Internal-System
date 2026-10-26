// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"accounting-be/api"
	"github.com/joho/godotenv"
)

func main() {
	// Muat environment variables dari file .env (jika ada)
	if err := godotenv.Load(); err != nil {
		fmt.Println("Catatan: File .env tidak ditemukan, menggunakan nilai default sistem.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server Backend berjalan di http://localhost:%s\n", port)
	fmt.Printf("Test API Ping  : http://localhost:%s/api/ping\n", port)
	fmt.Printf("Test API Banks : http://localhost:%s/api/banks\n", port)
	fmt.Printf("Test API Accts : http://localhost:%s/api/accounts\n", port)

	http.HandleFunc("/", api.Handler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Gagal menjalankan server: %v\n", err)
	}
}