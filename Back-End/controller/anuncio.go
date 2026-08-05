package controller

import (
	"Frota/db"
	"Frota/structs"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ==========================================
// ROTAS DO ADMIN
// ==========================================

// CriarAnuncio (POST) - Recebe o MultipartForm com a foto e os textos
func CriarAnuncio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro": "Método inválido"}`, http.StatusMethodNotAllowed)
		return
	}

	// Limita o tamanho do upload na memória (5MB)
	r.ParseMultipartForm(5 << 20)

	empresa := r.FormValue("empresa")
	linkDestino := r.FormValue("link_destino")

	if empresa == "" || linkDestino == "" {
		http.Error(w, `{"erro": "Campos obrigatórios faltando"}`, http.StatusBadRequest)
		return
	}

	// Processa o arquivo de imagem
	file, handler, err := r.FormFile("imagem")
	if err != nil {
		http.Error(w, `{"erro": "Imagem não enviada"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Cria a pasta de destino se não existir
	pastaUploads := "./uploads/anuncios"
	os.MkdirAll(pastaUploads, os.ModePerm)

	// Gera um nome único para o arquivo (evitar substituição de arquivos com o mesmo nome)
	nomeArquivo := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(handler.Filename))
	caminhoCompleto := filepath.Join(pastaUploads, nomeArquivo)

	// Salva o arquivo no disco do servidor
	dst, err := os.Create(caminhoCompleto)
	if err != nil {
		http.Error(w, `{"erro": "Erro ao salvar imagem"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Salva no Banco de Dados
	anuncio := structs.Anuncio{
		Empresa:     empresa,
		LinkDestino: linkDestino,
		ImagemURL:   "/uploads/anuncios/" + nomeArquivo, // Caminho público
		Ativo:       true,
		Cliques:     0,
		CreatedAt:   time.Now(),
	}

	if err := db.DB.Create(&anuncio).Error; err != nil {
		http.Error(w, `{"erro": "Falha ao salvar no banco"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
		"anuncio": anuncio,
	})
}

// ListarAnunciosAdmin (GET) - Mostra todos, ativos e pausados
func ListarAnunciosAdmin(w http.ResponseWriter, r *http.Request) {
	var anuncios []structs.Anuncio
	db.DB.Order("created_at desc").Find(&anuncios)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anuncios)
}

// ToggleAnuncio (PUT) - Ativa ou Pausa um anúncio
func ToggleAnuncio(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID da URL (Ex: /api/admin/anuncios/5/toggle)
	partes := strings.Split(r.URL.Path, "/")
	idStr := partes[len(partes)-2]

	var anuncio structs.Anuncio
	if err := db.DB.First(&anuncio, idStr).Error; err != nil {
		http.Error(w, `{"erro": "Anúncio não encontrado"}`, http.StatusNotFound)
		return
	}

	// Inverte o status
	novoStatus := !anuncio.Ativo
	db.DB.Model(&anuncio).Update("ativo", novoStatus)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"sucesso": true, "novo_status": novoStatus})
}

// DeletarAnuncio (DELETE)
func DeletarAnuncio(w http.ResponseWriter, r *http.Request) {
	partes := strings.Split(r.URL.Path, "/")
	idStr := partes[len(partes)-1]

	var anuncio structs.Anuncio
	if err := db.DB.First(&anuncio, idStr).Error; err == nil {
		// Remove o arquivo físico de imagem do disco para economizar espaço
		caminhoArquivo := "." + anuncio.ImagemURL
		os.Remove(caminhoArquivo)

		// Apaga do banco
		db.DB.Delete(&anuncio)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
}

// ==========================================
// ROTAS DO PASSAGEIRO
// ==========================================

// ListarAnunciosAtivos (GET) - Entrega apenas os que estão rodando
func ListarAnunciosAtivos(w http.ResponseWriter, r *http.Request) {
	var anuncios []structs.Anuncio
	db.DB.Where("ativo = ?", true).Order("created_at desc").Find(&anuncios)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anuncios)
}

// RegistrarClique (POST) - Adiciona +1 clique invisivelmente
func RegistrarClique(w http.ResponseWriter, r *http.Request) {
	partes := strings.Split(r.URL.Path, "/")
	idStr := partes[len(partes)-2] // Ex: /api/anuncios/5/clique

	// O GORM soma de forma atômica no banco!
	db.DB.Model(&structs.Anuncio{}).Where("id = ?", idStr).UpdateColumn("cliques", gorm.Expr("cliques + ?", 1))

	w.WriteHeader(http.StatusOK)
}
