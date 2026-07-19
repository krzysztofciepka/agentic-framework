package api

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var imageExts = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "TOO_LARGE", "max 10MB")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FILE", "image field required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	ext, ok := imageExts[contentType]
	if !ok {
		writeError(w, http.StatusBadRequest, "INVALID_TYPE", "only PNG, JPEG, GIF, WebP")
		return
	}

	id := make([]byte, 16)
	rand.Read(id)
	filename := hex.EncodeToString(id) + ext

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "data/uploads"
	}
	os.MkdirAll(uploadDir, 0755)

	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":  filename,
		"url": "/uploads/" + filename,
	})
}
