package main

import (
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"scos-lending/pkg"
)

func main() {
	server := pkg.NewServer()

	if err := server.InitClients(); err != nil {
		log.Fatal("Failed to initialize clients:", err)
	}

	// 启动价格监控
	go server.StartPriceMonitoring()

	// 设置路由
	router := server.SetupRoutes()

	// CORS支持
	corsObj := handlers.AllowedOrigins([]string{"*"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	log.Printf("Server starting on port %s", server.GetPort())
	log.Fatal(http.ListenAndServe(":"+server.GetPort(), handlers.CORS(corsObj, corsHeaders, corsMethods)(router)))
}
