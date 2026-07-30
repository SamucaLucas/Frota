package controller

import (
	"Frota/db"
	"Frota/structs"
	"encoding/json"
	"net/http"
)

func ObterDetalhesCorrida(w http.ResponseWriter, r *http.Request) {
	corridaID := r.URL.Query().Get("id")
	if corridaID == "" {
		http.Error(w, `{"erro": "ID da corrida não informado"}`, http.StatusBadRequest)
		return
	}

	var corrida structs.LocalizacaoMotorista // Substitua pela sua struct real de Corrida Fechada/Histórico
	// Fazemos os Joins para pegar o nome do motorista e do passageiro
	err := db.DB.Preload("Usuario").Preload("Motorista").First(&corrida, "id = ?", corridaID).Error

	if err != nil {
		http.Error(w, `{"erro": "Corrida não encontrada"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(corrida)
}
