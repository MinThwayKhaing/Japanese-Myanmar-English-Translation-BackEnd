package handlers

import (
	"USDT_BackEnd/models"
	"USDT_BackEnd/services"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordHandler struct {
	service     *services.WordService
	userService *services.UserService
}

func NewWordHandler(userService *services.UserService) *WordHandler {
	return &WordHandler{
		service:     services.NewWordService(),
		userService: userService,
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
	if err != nil || existingWord == nil {
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
	word, err := h.service.GetWordByID(r.Context(), id)
	if err == nil && word != nil && word.ImageURL != "" {
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
	if err != nil || word == nil {
		http.Error(w, "Word not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(word)
}

func (h *WordHandler) SearchWords(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	query := r.URL.Query().Get("q")
	words, err := h.service.SearchWords(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(words)
}

func (h *WordHandler) GetWordByID(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	// Check and decrement search limit
	if err := h.userService.CheckAndDecrementSearches(r.Context(), userID); err != nil {
		if err.Error() == "SEARCH_LIMIT_REACHED" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"code":    "SEARCH_LIMIT_REACHED",
				"message": "Usage limit reached",
			})
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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

// ------------------ DUPLICATE SYNC ------------------

func (h *WordHandler) GetDuplicateWords(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	query := r.URL.Query().Get("q")

	groups, totalGroups, err := h.service.GetDuplicateWords(r.Context(), page, limit, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	hasMore := int64(page*limit) < totalGroups

	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups":      groups,
		"totalGroups": totalGroups,
		"hasMore":     hasMore,
		"currentPage": page,
	})
}

// ------------------ IGNORE TOGGLE ------------------

func (h *WordHandler) SetWordIgnore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WordID string `json:"wordId"`
		Ignore bool   `json:"ignore"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.WordID == "" {
		http.Error(w, "wordId is required", http.StatusBadRequest)
		return
	}

	if err := h.service.SetWordIgnore(r.Context(), req.WordID, req.Ignore); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update word: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Word updated",
		"ignore":  req.Ignore,
	})
}

// ------------------ EXCEL UPLOAD ------------------

func (h *WordHandler) ExcelCreateWords(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200 << 20) // 200MB for large Excel files
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
		http.Error(w, fmt.Sprintf("Invalid Excel file: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("Parsed %d words from Excel, starting bulk insert...\n", len(words))

	// Use a longer timeout for large inserts (10 minutes)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	inserted, err := h.service.BulkCreateWords(ctx, words)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save words (inserted %d/%d): %v", inserted, len(words), err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Bulk insert complete: %d words inserted\n", inserted)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  fmt.Sprintf("%d words inserted successfully", inserted),
		"total":    inserted,
	})
}

// Excel parsing helper — uses row iterator for memory efficiency with large files
func parseExcelFile(file multipart.File) ([]models.Word, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rowIter, err := f.Rows(sheet)
	if err != nil {
		return nil, err
	}
	defer rowIter.Close()

	words := make([]models.Word, 0, 1024)
	isHeader := true

	for rowIter.Next() {
		if isHeader {
			isHeader = false
			continue // skip header row
		}

		row, err := rowIter.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		// Ensure row has at least 4 columns (Kanji,Hiragana,English,Myanmar)
		for len(row) < 4 {
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
