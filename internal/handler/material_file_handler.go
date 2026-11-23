package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/yourorg/portal/internal/domain"
	"github.com/yourorg/portal/internal/handler/response"
)

const maxDetectBytes = 512

// POST /materials/{id}/file
func (h *MaterialHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		response.JSON(w, http.StatusRequestEntityTooLarge,
			map[string]string{"error": "файл слишком большой или форма битая"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "нужно поле file"})
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "пустой файл"})
		return
	}

	ext := strings.ToLower(strings.TrimPrefix(filepathExt(header.Filename), "."))
	if ext == "doc" || ext == "docx" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "DOCX не принимаем: сохраните как PDF или создайте текстовый материал",
		})
		return
	}
	if ext != "pdf" {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "нужен файл с расширением .pdf"})
		return
	}

	head := make([]byte, maxDetectBytes)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "не удалось прочитать файл"})
		return
	}
	trimmed := bytes.TrimLeft(head[:n], "\x00\t\r\n ")
	if !bytes.HasPrefix(trimmed, []byte("%PDF-")) || !strings.HasPrefix(http.DetectContentType(trimmed), "application/pdf") {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "содержимое не похоже на PDF"})
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "seek error"})
		return
	}

	key, materialID, err := h.uc.UploadFile(r.Context(), v, id, domain.MaterialTypePDF, file, header.Size)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"id":       materialID,
		"file_key": key,
		"type":     domain.MaterialTypePDF,
	})
}

// GET /materials/{id}/download
func (h *MaterialHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	v, ok := currentViewer(r)
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	url, err := h.uc.GetDownloadURL(r.Context(), v, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"url":        url,
		"expires_in": int(h.presignTTL.Seconds()),
	})
}

func filepathExt(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[i:]
	}
	return ""
}
