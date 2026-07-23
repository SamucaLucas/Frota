package structs

import (
	"time"

	"gorm.io/gorm"
)

// 1. Usuários (Passageiros, Motoristas e Admin)
type Usuario struct {
	ID            uint   `gorm:"primaryKey"`
	Nome          string `gorm:"size:255;not null"`
	Email         string `gorm:"size:255;unique;not null"`
	Senha         string `gorm:"size:255;not null"`
	Whatsapp      string `gorm:"size:20;not null"`
	Papel         string `gorm:"size:20;not null"` // passageiro, motorista, admin
	Tokens        int    `gorm:"default:0"`        // Saldo atual para facilitar a visualização
	AceitouTermos bool   `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"` // Para conformidade com LGPD (Soft Delete)
}

func (Usuario) TableName() string {
	return "usuarios"
}

// 2. Corridas (O coração operacional)
type Corrida struct {
	ID               uint    `gorm:"primaryKey"`
	UsuarioID        uint    `gorm:"not null"` // Quem solicitou
	Usuario          Usuario `gorm:"foreignKey:UsuarioID"`
	MotoristaID      *uint   // Quem aceitou (Pode ser nulo até a aprovação)
	Motorista        Usuario `gorm:"foreignKey:MotoristaID"`
	Tipo             string  `gorm:"size:50;not null"` // padrao, livre
	DataHoraAgendada time.Time
	OrigemTexto      string `gorm:"not null"`
	OrigemLat        float64
	OrigemLng        float64
	DestinoTexto     string
	DestinoLat       float64
	DestinoLng       float64
	KMRodado         float64
	ValorEstimado    float64
	ValorFinal       float64
	Status           string `gorm:"size:50;default:'Aguardando Confirmacao'"` // Aguardando, Aprovada, Concluida, Cancelada
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (Corrida) TableName() string {
	return "corridas"
}

// 3. Recibos (Documentos Físicos Digitais)
type Recibo struct {
	ID          uint    `gorm:"primaryKey"`
	CorridaID   uint    `gorm:"not null"`
	Corrida     Corrida `gorm:"foreignKey:CorridaID"`
	UrlPdf      string  `gorm:"size:500;not null"`
	CnpjEmissor string  `gorm:"size:20"`
	CarroPlaca  string  `gorm:"size:20"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (Recibo) TableName() string {
	return "recibos"
}

// 4. Histórico de Tokens (O Ledger para rastreabilidade)
type HistoricoToken struct {
	ID            uint    `gorm:"primaryKey"`
	UsuarioID     uint    `gorm:"not null"`
	Usuario       Usuario `gorm:"foreignKey:UsuarioID"`
	Quantidade    int     `gorm:"not null"`         // Ex: +1, +5, -1
	TipoTransacao string  `gorm:"size:50;not null"` // corrida, compra_avulsa, uso_sorteio
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (HistoricoToken) TableName() string {
	return "historico_tokens"
}

// 5. Campanhas (A Vitrine do Sorteio)
type Campanha struct {
	ID              uint    `gorm:"primaryKey"`
	Titulo          string  `gorm:"size:255;not null"`
	PremioNome      string  `gorm:"size:255;not null"`
	PremioImagemUrl string  `gorm:"size:500"`
	Status          string  `gorm:"size:50;default:'Ativa'"` // Ativa, Finalizada
	GanhadorID      *uint   // Fica nulo até o Dudu sortear
	Ganhador        Usuario `gorm:"foreignKey:GanhadorID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Campanha) TableName() string {
	return "campanhas"
}

// 6. Números do Sorteio (A Grade de 1 a 1000)
type NumeroSorteio struct {
	ID         uint     `gorm:"primaryKey"`
	CampanhaID uint     `gorm:"not null"`
	Campanha   Campanha `gorm:"foreignKey:CampanhaID"`
	Numero     int      `gorm:"not null"`                // 1 a 1000
	Status     string   `gorm:"size:20;default:'Livre'"` // Livre (Verde), Ocupado (Vermelho)
	UsuarioID  *uint    // Quem pegou o número
	Usuario    Usuario  `gorm:"foreignKey:UsuarioID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (NumeroSorteio) TableName() string {
	return "numeros_sorteio"
}

// 7. Preços e Rotas (Configurações Financeiras Administrativas)
type PrecoRota struct {
	ID           uint    `gorm:"primaryKey"`
	TipoCobranca string  `gorm:"size:50;not null"` // Rota Fixa, Por KM
	OrigemBase   string  `gorm:"size:255"`
	DestinoBase  string  `gorm:"size:255"`
	Valor        float64 `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (PrecoRota) TableName() string {
	return "precos_rotas"
}

// 8. Patrocínios (Módulo de Monetização)
type Patrocinio struct {
	ID              uint   `gorm:"primaryKey"`
	Titulo          string `gorm:"size:255;not null"`
	ImagemBannerUrl string `gorm:"size:500;not null"`
	LinkDestino     string `gorm:"size:500"`
	Status          string `gorm:"size:50;default:'Ativo'"` // Ativo, Inativo
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Patrocinio) TableName() string {
	return "patrocinios"
}

// 9. Transações Financeiras (O Livro Caixa exclusivo para Entradas)
type TransacaoFinanceira struct {
	ID        uint    `gorm:"primaryKey"`
	CorridaID *uint   // Pode ser nulo se for venda avulsa ou patrocínio
	Corrida   Corrida `gorm:"foreignKey:CorridaID"`
	UsuarioID *uint   // Quem pagou
	Usuario   Usuario `gorm:"foreignKey:UsuarioID"`
	Categoria string  `gorm:"size:100;not null"` // Corrida, Venda Token, Patrocinio
	Valor     float64 `gorm:"not null"`
	Descricao string  `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (TransacaoFinanceira) TableName() string {
	return "transacoes_financeiras"
}
