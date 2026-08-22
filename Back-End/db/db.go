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
		&structs.Veiculo{},
		&structs.Corrida{},
		&structs.Recibo{},
		&structs.HistoricoToken{},
		&structs.Campanha{},
		&structs.NumeroSorteio{},
		&structs.PrecoRota{},
		&structs.Patrocinio{},
		&structs.TransacaoFinanceira{},
		&structs.LocalizacaoMotorista{},
		&structs.ConfiguracaoApp{},
		&structs.Anuncio{},
	)

	if err != nil {
		log.Fatal("❌ Erro ao rodar as migrações (AutoMigrate):\n", err)
	}

	fmt.Println("🚀 Todas as tabelas (incluindo Veiculos) foram criadas/sincronizadas com sucesso!")

	// 5. Atribuindo a conexão aberta à nossa variável global
	DB = database

	// 6. Garante a criação do Passageiro Avulso e das Triggers de proteção
	inicializarPassageiroAvulsoETriggers(database)
}

func inicializarPassageiroAvulsoETriggers(database *gorm.DB) {
	// 1. Verifica se o Passageiro Avulso já existe
	var count int64
	database.Model(&structs.Usuario{}).Where("email = ? OR papel = ?", "Passageiro@LT.com", "sistema").Count(&count)
	if count == 0 {
		var countID1 int64
		database.Model(&structs.Usuario{}).Where("id = ?", 1).Count(&countID1)

		passageiroAvulso := structs.Usuario{
			Nome:          "Passageiro Avulso",
			Email:         "Passageiro@LT.com",
			Senha:         "",
			Whatsapp:      "000000000000",
			Papel:         "sistema",
			AceitouTermos: true,
		}

		if countID1 == 0 {
			passageiroAvulso.ID = 1
		}

		if err := database.Create(&passageiroAvulso).Error; err != nil {
			log.Println("⚠️ Aviso: Não foi possível criar o Passageiro Avulso:", err)
		} else {
			fmt.Printf("👤 Passageiro Avulso criado com sucesso (ID: %d)!\n", passageiroAvulso.ID)
		}
	}

	// 2. Cria/Atualiza a Função PL/pgSQL no PostgreSQL
	queryFuncao := `
CREATE OR REPLACE FUNCTION proteger_passageiro_avulso()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.id = 1 OR OLD.email = 'Passageiro@LT.com' OR OLD.papel = 'sistema' THEN
        RAISE EXCEPTION 'Acesso Negado pelo Banco: O Passageiro Avulso é uma engrenagem do sistema e não pode ser editado nem excluído!';
    END IF;

    IF TG_OP = 'UPDATE' THEN
        RETURN NEW;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
`
	if err := database.Exec(queryFuncao).Error; err != nil {
		log.Println("⚠️ Erro ao criar a função proteger_passageiro_avulso:", err)
	}

	// 3. Cria as Triggers de DELETE e UPDATE no PostgreSQL se não existirem
	queryTriggers := `
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_bloquear_delete_avulso') THEN
        CREATE TRIGGER trg_bloquear_delete_avulso
        BEFORE DELETE ON usuarios
        FOR EACH ROW
        EXECUTE PROCEDURE proteger_passageiro_avulso();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_bloquear_update_avulso') THEN
        CREATE TRIGGER trg_bloquear_update_avulso
        BEFORE UPDATE ON usuarios
        FOR EACH ROW
        EXECUTE PROCEDURE proteger_passageiro_avulso();
    END IF;
END $$;
`
	if err := database.Exec(queryTriggers).Error; err != nil {
		log.Println("⚠️ Erro ao criar triggers no banco:", err)
	} else {
		fmt.Println("🔒 Triggers de proteção do Passageiro Avulso sincronizadas com sucesso!")
	}
}

