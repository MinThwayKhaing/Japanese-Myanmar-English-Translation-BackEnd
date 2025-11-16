package handlers

import (
	"USDT_BackEnd/models"
	"USDT_BackEnd/services"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type WordHandler struct {
	service *services.WordService
}

func NewWordHandler() *WordHandler {
	return &WordHandler{
		service: services.NewWordService(),
	}
}

// ------------------ WORD CRUD ------------------

func (h *WordHandler) CreateWord(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	word := models.Word{
		Japanese: r.FormValue("japanese"),
		SubTerm:  r.FormValue("subTerm"),
		English:  r.FormValue("english"),
		Myanmar:  r.FormValue("myanmar"),
	}

	// Handle image
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		storage := services.NewStorageService()
		fileName := fmt.Sprintf("words/%d_%s", time.Now().Unix(), header.Filename)
		imageURL, err := storage.UploadFile(file, fileName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to upload image: %v", err), http.StatusInternalServerError)
			return
		}
		word.ImageURL = imageURL
	}

	err = h.service.CreateWord(r.Context(), &word)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create word: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(word)
}
func (h *WordHandler) UpdateWord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get existing word (to know old image URL)
	existingWord, err := h.service.GetWordByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Word not found", http.StatusNotFound)
		return
	}

	word := models.Word{
		Japanese: r.FormValue("japanese"),
		SubTerm:  r.FormValue("subTerm"),
		English:  r.FormValue("english"),
		Myanmar:  r.FormValue("myanmar"),
	}

	storage := services.NewStorageService()

	// Handle image update
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// 1️⃣ Delete old image if exists
		if existingWord.ImageURL != "" {
			oldFileName := existingWord.ImageURL[strings.LastIndex(existingWord.ImageURL, "/")+1:]
			_ = storage.DeleteFile("words/" + oldFileName)
		}

		// 2️⃣ Upload new image
		fileName := fmt.Sprintf("words/%d_%s", time.Now().Unix(), header.Filename)
		imageURL, err := storage.UploadFile(file, fileName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to upload new image: %v", err), http.StatusInternalServerError)
			return
		}
		word.ImageURL = imageURL
	} else {
		// Keep existing image if no new image uploaded
		word.ImageURL = existingWord.ImageURL
	}

	// 3️⃣ Save to DB
	err = h.service.UpdateWord(r.Context(), id, &word)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update word: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(word)
}

func (h *WordHandler) DeleteWord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Optional: delete image from DO
	word, _ := h.service.GetWordByID(r.Context(), id)
	if word.ImageURL != "" {
		storage := services.NewStorageService()
		// Extract filename from URL
		fileName := word.ImageURL[strings.LastIndex(word.ImageURL, "/")+1:]
		storage.DeleteFile("words/" + fileName)
	}

	if err := h.service.DeleteWord(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete word: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ------------------ SELECT/SEARCH ------------------

func (h *WordHandler) SelectOneWord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	word, err := h.service.GetWordByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Word not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(word)
}

func (h *WordHandler) SearchWords(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	words, err := h.service.SearchWords(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(words)
}

func (h *WordHandler) GetWordByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	word, err := h.service.GetWordByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Word not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(word)
}

func (h *WordHandler) GetAllWords(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 15
	}
	query := r.URL.Query().Get("q")

	words, hasMore, totalCount, err := h.service.GetAllWords(r.Context(), page, limit, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type WordResponse struct {
		ID        string `json:"_id"`
		Japanese  string `json:"japanese,omitempty"`
		SubTerm   string `json:"subTerm,omitempty"`
		English   string `json:"english,omitempty"`
		Myanmar   string `json:"myanmar,omitempty"`
		ImageURL  string `json:"imageURL,omitempty"`
		CreatedAt string `json:"createdAt,omitempty"`
		UpdatedAt string `json:"updatedAt,omitempty"`
	}

	respWords := make([]WordResponse, len(words))
	for i, w := range words {
		respWords[i] = WordResponse{
			ID:        w.ID.Hex(),
			Japanese:  w.Japanese,
			SubTerm:   w.SubTerm,
			English:   w.English,
			Myanmar:   w.Myanmar,
			ImageURL:  w.ImageURL,
			CreatedAt: w.CreatedAt.Format(time.RFC3339),
			UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
		}
	}

	response := map[string]interface{}{
		"words":       respWords,
		"hasMore":     hasMore,
		"currentPage": page,
		"totalCount":  totalCount,
	}

	json.NewEncoder(w).Encode(response)
}

// ------------------ EXCEL UPLOAD ------------------

func (h *WordHandler) ExcelCreateWords(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20) // 20MB for Excel
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Println("Uploading Excel:", header.Filename)

	words, err := parseExcelFile(file)
	if err != nil {
		http.Error(w, "Invalid Excel file", http.StatusBadRequest)
		return
	}

	err = h.service.BulkCreateWords(r.Context(), words)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save words: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("%d words inserted successfully", len(words)),
	})
}

// Excel parsing helper
func parseExcelFile(file multipart.File) ([]models.Word, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	var words []models.Word
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}

		// Ensure row has at least 5 columns (Kanji,Hiragana,English,Myanmar,ignore extra)
		for len(row) < 5 {
			row = append(row, "")
		}

		word := models.Word{
			Japanese: strings.TrimSpace(row[0]), // Kanji
			SubTerm:  strings.TrimSpace(row[1]), // Hiragana
			English:  strings.TrimSpace(row[2]), // English
			Myanmar:  strings.TrimSpace(row[3]), // Myanmar
		}
		words = append(words, word)
	}

	return words, nil
}
