package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// "Frota/config"
	"Frota/controller"
	"Frota/db"
	"Frota/routers"
	"Frota/structs"

	"github.com/joho/godotenv"
)

// corsMiddleware é o nosso "porteiro" que libera o acesso do App (Capacitor) para a API Go
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Libera acesso de qualquer origem (O "*" permite o app do celular acessar)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Define quais métodos o celular pode usar
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		// Define quais cabeçalhos o celular pode enviar (importante para enviar JSON e Token)
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Se for uma requisição OPTIONS (Preflight do navegador/Capacitor), apenas devolve OK 200
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Se não for OPTIONS, passa a requisição adiante para as suas rotas reais
		next.ServeHTTP(w, r)
	})
}

func IniciarMonitoramentoMotoristas() {
	// Cria um "relógio" que bate a cada 1 minuto
	ticker := time.NewTicker(1 * time.Minute)

	// Cria uma Goroutine (uma thread rodando em segundo plano separada da API)
	go func() {
		for {
			<-ticker.C // Espera o relógio bater

			// Define o tempo limite: Agora menos 2 minutos
			tempoLimite := time.Now().Add(-2 * time.Minute)

			// Atualiza no banco: Quem não manda sinal há mais de 2 minutos e ainda está "Disponível", vira "Indisponível"
			// Substitua "SuaTabelaDeLocalizacao" pela tabela real onde você salva o GPS (ex: models.LocalizacaoMotorista)
			resultado := db.DB.Model(&structs.LocalizacaoMotorista{}).
				Where("updated_at < ? AND status = ?", tempoLimite, "Disponível").
				Update("status", "Indisponível")

			if resultado.Error != nil {
				log.Println("Erro ao varrer motoristas inativos:", resultado.Error)
			} else if resultado.RowsAffected > 0 {
				log.Printf("🧹 Varredor: %d motoristas inativos foram marcados como Indisponíveis.\n", resultado.RowsAffected)
			}
		}
	}()
}

func main() {
	// 1. Carregar as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Aviso: Arquivo .env não encontrado. O sistema tentará usar as variáveis nativas do SO.")
	}

	controller.ConfigurarGoogleOAuth()

	// 2. Inicializar a conexão com o PostgreSQL e rodar o AutoMigrate
	fmt.Println("⏳ Iniciando o Sistema de Frota - Dudu...")
	db.ConectarBanco()

	// 3. Aqui carregaremos as rotas (pasta routers)
	r := routers.ConfigurarRotas()

	

	IniciarMonitoramentoMotoristas()

	// Subindo o servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Notei que no seu teste anterior estava 8082! Ajuste se necessário.
	}

	fmt.Println("🚀 Servidor da Frota rodando na porta:", port)

	// 3. Inicia o servidor com o CORS ativado
	err = http.ListenAndServe("0.0.0.0:"+port, corsMiddleware(r))
	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}

}
