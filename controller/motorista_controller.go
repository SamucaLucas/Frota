package controller

import (
	"log"
	"net/http"

	"Frota/db"
	"Frota/services"
	"Frota/structs"
)

type DadosHomeMotorista struct {
	Usuario   structs.Usuario
	Pendentes []structs.Corrida
	Aprovadas []structs.Corrida
}

func HomeMotorista(w http.ResponseWriter, r *http.Request) {
	// 1. Quem é o motorista logado?
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Busca os dados dele
	var usuario structs.Usuario
	db.DB.First(&usuario, usuarioID)

	// Proteção: Se um passageiro tentar acessar a url do motorista, expulsa!
	if usuario.Papel != "motorista" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	// 2. Busca todas as corridas que os passageiros pediram e ninguém aceitou ainda
	var pendentes []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ?", "Aguardando Confirmacao").Order("data_hora_agendada ASC").Find(&pendentes)

	// 3. Busca as corridas que este motorista já aceitou
	var aprovadas []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ? AND motorista_id = ?", "Aprovada", usuarioID).Order("data_hora_agendada ASC").Find(&aprovadas)

	dados := DadosHomeMotorista{
		Usuario:   usuario,
		Pendentes: pendentes,
		Aprovadas: aprovadas,
	}

	err = temp.ExecuteTemplate(w, "MotoristaHome", dados)
	if err != nil {
		log.Println("Erro na renderização:", err)
	}
}
