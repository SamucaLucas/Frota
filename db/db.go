package db

import (
	"fmt"
	"log"
	"os"

	"Frota/structs" // Importando o pacote com as nossas entidades

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB é a variável global que vai segurar a nossa conexão com o banco
// Ela começa com letra maiúscula para poder ser exportada e usada em outros pacotes (como os controllers)
var DB *gorm.DB

func ConectarBanco() {
	// 1. Lendo as credenciais do nosso arquivo .env
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// 2. Montando a string de conexão (DSN - Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		host, user, password, dbname, port)

	// 3. Abrindo a conexão com o PostgreSQL através do GORM
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Erro fatal: Não foi possível conectar ao banco de dados!\n", err)
	}

	fmt.Println("✅ Conexão com o PostgreSQL estabelecida com sucesso!")

	// 4. A Mágica do Code First: AutoMigrate
	// O GORM vai ler nossas structs e criar/atualizar as tabelas no banco automaticamente
	err = database.AutoMigrate(
		&structs.Usuario{},
		&structs.Corrida{},
		&structs.Recibo{},
		&structs.HistoricoToken{},
		&structs.Campanha{},
		&structs.NumeroSorteio{},
		&structs.PrecoRota{},
		&structs.Patrocinio{},
		&structs.TransacaoFinanceira{},
	)

	if err != nil {
		log.Fatal("❌ Erro ao rodar as migrações (AutoMigrate):\n", err)
	}

	fmt.Println("🚀 Todas as 9 tabelas foram criadas/sincronizadas com sucesso!")

	// 5. Atribuindo a conexão aberta à nossa variável global
	DB = database
}
