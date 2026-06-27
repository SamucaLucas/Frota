package controller

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"Frota/db"
	"Frota/services"
	"Frota/structs"
)

// Pacote de dados que será enviado para a tela do Dudu
type DadosHomeAdmin struct {
	Usuario    structs.Usuario
	Pendentes  []structs.Corrida
	Aprovadas  []structs.Corrida
	Motoristas []structs.Usuario // Lista de todos os motoristas para o Dudu escolher
}

type DadosDespacho struct {
	Usuario    structs.Usuario
	Corrida    structs.Corrida
	Motoristas []structs.Usuario
}

func HomeAdmin(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var usuario structs.Usuario
	db.DB.First(&usuario, usuarioID)

	// Proteção: Apenas o ADMIN (Dudu) pode acessar essa tela
	if usuario.Papel != "admin" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	// 1. Busca todos os motoristas cadastrados na empresa
	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	// 2. Busca solicitações aguardando atribuição
	var pendentes []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ?", "Aguardando Confirmacao").Order("data_hora_agendada ASC").Find(&pendentes)

	// 3. Busca a agenda geral da frota (Corridas já atribuídas a algum motorista)
	var aprovadas []structs.Corrida
	db.DB.Preload("Usuario").Preload("Motorista").Where("status = ?", "Aprovada").Order("data_hora_agendada ASC").Find(&aprovadas)

	dados := DadosHomeAdmin{
		Usuario:    usuario,
		Pendentes:  pendentes,
		Aprovadas:  aprovadas,
		Motoristas: motoristas,
	}

	err = temp.ExecuteTemplate(w, "AdminHome", dados)
	if err != nil {
		log.Println("Erro na renderização do Admin:", err)
	}
}

func DespacharCorridaTela(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var admin structs.Usuario
	db.DB.First(&admin, usuarioID)

	if admin.Papel != "admin" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	// 1. Pega o ID da corrida na URL
	vars := mux.Vars(r)
	corridaID := vars["id"]

	// 2. Busca a corrida específica com os dados do passageiro
	var corrida structs.Corrida
	db.DB.Preload("Usuario").First(&corrida, corridaID)

	// 3. Busca os motoristas
	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	dados := DadosDespacho{
		Usuario:    admin,
		Corrida:    corrida,
		Motoristas: motoristas,
	}

	err = temp.ExecuteTemplate(w, "AdminDespachar", dados)
	if err != nil {
		log.Println("Erro na renderização da tela de Despacho:", err)
	}
}

// DespacharCorrida centraliza a exibição (GET) e o processamento (POST)
func DespacharCorrida(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var admin structs.Usuario
	db.DB.First(&admin, usuarioID)

	if admin.Papel != "admin" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	// Como usamos o ServeMux nativo, extraímos o ID "cortando" a rota base
	corridaID := strings.TrimPrefix(r.URL.Path, "/admin/despachar/")

	// ═════════════════════════════════════════════════════════════
	// CONDIÇÃO 1: MÉTODO GET (Mostrar o Ecrã de Despacho)
	// ═════════════════════════════════════════════════════════════
	if r.Method == http.MethodGet {
		var corrida structs.Corrida
		db.DB.Preload("Usuario").First(&corrida, corridaID)

		var motoristas []structs.Usuario
		db.DB.Where("papel = ?", "motorista").Find(&motoristas)

		dados := struct {
			Usuario    structs.Usuario
			Corrida    structs.Corrida
			Motoristas []structs.Usuario
		}{
			Usuario:    admin,
			Corrida:    corrida,
			Motoristas: motoristas,
		}

		err = temp.ExecuteTemplate(w, "AdminDespachar", dados)
		if err != nil {
			log.Println("Erro ao renderizar o ecrã de Despacho:", err)
		}
		return
	}

	// ═════════════════════════════════════════════════════════════
	// CONDIÇÃO 2: MÉTODO POST (Salvar a atribuição no Banco)
	// ═════════════════════════════════════════════════════════════
	if r.Method == http.MethodPost {
		motoristaID := r.FormValue("motorista_id")

		// TRAVA DE SEGURANÇA: Impede que guarde vazio no banco de dados
		if motoristaID == "" {
			log.Println("Aviso: Tentativa de despacho sem motorista selecionado.")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther) // Recarrega a página
			return
		}

		errDb := db.DB.Model(&structs.Corrida{}).Where("id = ?", corridaID).Updates(map[string]interface{}{
			"motorista_id": motoristaID,
			"status":       "Aprovada",
		}).Error

		if errDb != nil {
			log.Println("Erro ao salvar atribuição no banco:", errDb)
		}

		// Volta para a Home do Admin
		http.Redirect(w, r, "/admin/home", http.StatusSeeOther)
		return
	}
}
