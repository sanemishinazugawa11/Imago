package main

import (
	"fmt"
	"image/png"
	"io"
	"strings"

	// "mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

var imageMap = make(map[string]string)

func home(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Welcome to our application"))
}

// handler for compressing the images
func compressHandler(w http.ResponseWriter, r *http.Request) {
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

	// Compression level from query
	level := r.URL.Query().Get("level")
	lev, err := strconv.Atoi(level)
	if err != nil {
		http.Error(w, "Invalid compression level", http.StatusBadRequest)
		return
	}

	compressedId := uuid.NewMD5(uuid.NameSpaceOID, []byte(id.String()))
	comPath := "./uploads/" + compressedId.String() + ext

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

	// Stream compressed file back
	w.Header().Set("Content-Disposition", "attachment; filename="+header.Filename)
	w.Header().Set("Content-Type", "image/"+strings.TrimPrefix(ext, "."))
	http.ServeFile(w, r, comPath)
}

// handler for downloading modified images
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Missing file parameter", http.StatusBadRequest)
		return
	}

	fpath := filepath.Join("uploads", fileName)
	http.ServeFile(w, r, fpath)
}

// handler for converting images from one format to another
func convertHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {

	http.HandleFunc("/", home)
	// http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/compress", compressHandler)
	http.HandleFunc("/download", downloadHandler)
	fmt.Println("Server running successfully")
	http.ListenAndServe(":8000", nil)
}
