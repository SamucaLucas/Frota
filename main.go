package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// "Frota/config"
	"Frota/db"

	"github.com/joho/godotenv"
	// "Frota/routers"
)

func main() {
	// 1. Carregar as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Aviso: Arquivo .env não encontrado. O sistema tentará usar as variáveis nativas do SO.")
	}

	// 2. Inicializar a conexão com o PostgreSQL e rodar o AutoMigrate
	fmt.Println("⏳ Iniciando o Sistema de Frota - Dudu...")
	db.ConectarBanco()

	// 3. Aqui carregaremos as rotas (pasta routers)

	// Subindo o servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback caso o .env falhe
	}

	fmt.Println("🚀 Servidor Frota rodando: http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("❌ Erro fatal ao iniciar o servidor: ", err)
	}
}
