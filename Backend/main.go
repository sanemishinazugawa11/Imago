package main

import (
	"fmt"
	"image/png"
	"io"
	"strings"

	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to our application"))
}

// handler for compressing the images
func compressHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Please upload file below 10 MB", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, "Error reading the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// unique ID for every image
	id, err := uuid.NewUUID()
	if err != nil {
		http.Error(w, "Error generating ID", http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		http.Error(w, "File has no extension", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("./uploads", os.FileMode(os.O_RDWR))
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}
	origPath := "./uploads/" + id.String() + ext
	dst, err := os.Create(origPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	img, err := imaging.Open(origPath)
	if err != nil {
		http.Error(w, "Error opening image", http.StatusInternalServerError)
		return
	}

	level := r.URL.Query().Get("level")
	lev, err := strconv.Atoi(level)
	if err != nil {
		http.Error(w, "Invalid compression level", http.StatusBadRequest)
		return
	}

	compressedId := uuid.NewMD5(uuid.NameSpaceOID, []byte(id.String()))
	err = os.MkdirAll("./uploads/processed", os.FileMode(os.O_RDWR))
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}
	comPath := "./uploads/processed/" + compressedId.String() + ext

	switch ext {
	case ".png":
		if lev < 0 {
			lev = 0
		}
		if lev > 3 {
			lev = 3
		}

		err = imaging.Save(img, comPath, imaging.PNGCompressionLevel(png.CompressionLevel(lev)))
	case ".jpg", ".jpeg":
		if lev < 1 {
			lev = 1
		}
		if lev > 100 {
			lev = 100
		}
		err = imaging.Save(img, comPath, imaging.JPEGQuality(lev))
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Error saving compressed image: "+err.Error(), http.StatusInternalServerError)
		return
	}
	relativePath := strings.TrimPrefix(comPath, "./uploads/")
	relativePath = strings.TrimPrefix(relativePath, ".\\uploads\\")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"file":"%s"}`, relativePath)
}

// handler for downloading modified images
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Missing file parameter", http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(fileName)
	fullPath := filepath.Join("uploads", cleanPath)

	normalizedPath := filepath.ToSlash(fullPath)

	if !strings.HasPrefix(normalizedPath, "uploads/") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", "attachment")
	http.ServeFile(w, r, fullPath)
}

// handler for converting images from one format to another
func convertHandler(w http.ResponseWriter, r *http.Request) {

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Please upload file below 10 MB", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, "Error reading the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate unique ID
	id, err := uuid.NewUUID()
	if err != nil {
		http.Error(w, "Error generating ID", http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		http.Error(w, "File has no extension", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("./uploads", os.FileMode(os.O_RDWR))
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	origPath := "./uploads/" + id.String() + ext
	dst, err := os.Create(origPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	img, err := imaging.Open(origPath)
	if err != nil {
		http.Error(w, "Error opening image", http.StatusInternalServerError)
		return
	}

	err = os.MkdirAll("./uploads/processed", os.FileMode(os.O_RDWR))
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	var DownloadPath string
	switch ext {

	case ".png":
		savingExt := ".jpeg"
		convertedId := uuid.NewMD5(uuid.NameSpaceOID, []byte(origPath))
		savingPath := "./uploads/processed/" + convertedId.String() + savingExt
		DownloadPath = savingPath
		err = imaging.Save(img, savingPath)
		if err != nil {
			http.Error(w, "Error saving converted image"+err.Error(), http.StatusInternalServerError)
			return
		}

	case ".jpg", ".jpeg":
		savingExt := ".png"
		convertedId := uuid.NewMD5(uuid.NameSpaceOID, []byte(origPath))
		savingPath := "./uploads/processed/" + convertedId.String() + savingExt
		DownloadPath = savingPath
		err = imaging.Save(img, savingPath)
		if err != nil {
			http.Error(w, "Error saving converted image"+err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}

	relativePath := strings.TrimPrefix(DownloadPath, "./uploads/")
	relativePath = strings.TrimPrefix(relativePath, ".\\uploads\\")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"file":"%s"}`, relativePath)
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	http.HandleFunc("/", corsMiddleware(home))
	http.HandleFunc("/compress", corsMiddleware(compressHandler))
	http.HandleFunc("/convert", corsMiddleware(convertHandler))
	http.HandleFunc("/download", corsMiddleware(downloadHandler))

	fmt.Println("Server running successfully on port 8000")
	http.ListenAndServe(":8000", nil)
}
