package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Frota/db"
	"Frota/structs"
)

// ObterPerfil retorna os dados do usuário logado e seus veículos
func ObterPerfil(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usuarioID := r.Context().Value("usuario_id").(uint)

	var usuario structs.Usuario
	if err := db.DB.First(&usuario, usuarioID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Usuário não encontrado."})
		return
	}

	var veiculos []structs.Veiculo
	if usuario.Papel == "motorista" || usuario.Papel == "admin" {
		db.DB.Where("motorista_id = ?", usuarioID).Order("ativo DESC, id ASC").Find(&veiculos)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":  true,
		"usuario":  usuario,
		"veiculos": veiculos,
	})
}

// AtualizarPerfil atualiza dados de texto
func AtualizarPerfil(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usuarioID := r.Context().Value("usuario_id").(uint)

	var req struct {
		Nome     string `json:"nome"`
		Email    string `json:"email"`
		Whatsapp string `json:"whatsapp"`
		Genero   string `json:"genero"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Formato inválido."})
		return
	}

	// Não vamos permitir atualizar senha por aqui ainda, apenas dados básicos
	db.DB.Model(&structs.Usuario{}).Where("id = ?", usuarioID).Updates(map[string]interface{}{
		"nome":     req.Nome,
		"email":    req.Email,
		"whatsapp": req.Whatsapp,
		"genero":   req.Genero,
	})

	json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": true})
}

// UploadFotoPerfil salva uma foto localmente
func UploadFotoPerfil(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usuarioID := r.Context().Value("usuario_id").(uint)

	err := r.ParseMultipartForm(5 << 20) // 5 MB max
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Arquivo muito grande ou inválido."})
		return
	}

	file, header, err := r.FormFile("foto")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao ler imagem."})
		return
	}
	defer file.Close()

	// Criar pasta se não existir
	pastaUploads := "./uploads/perfis"
	os.MkdirAll(pastaUploads, os.ModePerm)

	ext := strings.ToLower(filepath.Ext(header.Filename))
	novoNome := strconv.FormatUint(uint64(usuarioID), 10) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ext
	caminhoCompleto := filepath.Join(pastaUploads, novoNome)

	out, err := os.Create(caminhoCompleto)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao salvar imagem."})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao gravar arquivo."})
		return
	}

	// Salva o caminho no DB
	fotoURL := "/uploads/perfis/" + novoNome
	db.DB.Model(&structs.Usuario{}).Where("id = ?", usuarioID).Update("foto_perfil", fotoURL)

	json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": true, "foto": fotoURL})
}

// ==============================
// VEÍCULOS
// ==============================

// AdicionarVeiculo
func AdicionarVeiculo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usuarioID := r.Context().Value("usuario_id").(uint)

	var req structs.Veiculo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Dados inválidos."})
		return
	}

	// Valida campos
	if req.Placa == "" || req.Modelo == "" || req.Cor == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Placa, Modelo e Cor são obrigatórios."})
		return
	}

	// Se for o primeiro veículo, define como ativo automaticamente
	var count int64
	db.DB.Model(&structs.Veiculo{}).Where("motorista_id = ?", usuarioID).Count(&count)
	
	req.MotoristaID = usuarioID
	req.Ativo = (count == 0)

	if err := db.DB.Create(&req).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Falha ao salvar veículo."})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": true, "veiculo": req})
}

func extrairID(r *http.Request, prefixo string) string {
	return strings.TrimPrefix(r.URL.Path, prefixo)
}

// DefinirVeiculoAtivo
func DefinirVeiculoAtivo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usuarioID := r.Context().Value("usuario_id").(uint)
	veiculoID := extrairID(r, "/api/veiculos/ativo/")

	// Zera os ativos do motorista
	db.DB.Model(&structs.Veiculo{}).Where("motorista_id = ?", usuarioID).Update("ativo", false)

	// Seta apenas o selecionado
	err := db.DB.Model(&structs.Veiculo{}).Where("id = ? AND motorista_id = ?", veiculoID, usuarioID).Update("ativo", true).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao atualizar."})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": true})
}

// ExcluirVeiculo
func ExcluirVeiculo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usuarioID := r.Context().Value("usuario_id").(uint)
	veiculoID := extrairID(r, "/api/veiculos/excluir/")

	db.DB.Where("id = ? AND motorista_id = ?", veiculoID, usuarioID).Delete(&structs.Veiculo{})
	json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": true})
}
