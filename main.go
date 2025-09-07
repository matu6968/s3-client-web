package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"context"

	"github.com/joho/godotenv"
	"github.com/matu6968/s3-client/s3client"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/delete", handleDelete)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer file.Close()

	directory := r.FormValue("directory")
	if directory == "" {
		directory = "/"
	}

	tempFileName := filepath.Join(os.TempDir(), handler.Filename)
	tempFile, err := os.Create(tempFileName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	ctx := context.TODO()
	client, err := s3client.LoadClient(ctx, "", true)
	if err != nil {
		log.Fatal("Error initializing client:", err)
		return
	}
	output, err := client.UploadFile(ctx, tempFile.Name(), directory, true)
	if err != nil {
		log.Printf("Error uploading file: %s\n", err)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Error uploading file to S3: %s", err),
		})
		return
	}
	
	response := struct {
		Message string `json:"message"`
		Output  string `json:"output"`
	}{
		Message: fmt.Sprintf("File %s uploaded successfully", handler.Filename),
		Output:  string(fmt.Sprintf("Uploaded:", output)),
	}

	json.NewEncoder(w).Encode(response)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Filename is required"})
		return
	}
	ctx := context.TODO()
	client, err := s3client.LoadClient(ctx, "", true)
	if err != nil {
		log.Fatal("Error initializing client:", err)
		return
	}
	err = client.DeleteFile(ctx, filename)
	if err != nil {
		log.Printf("Error deleting file: %s\n", err)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Error deleting file from S3: %s", err),
		})
		return
	}

	response := struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf("File %s deleted successfully from S3", filename),
	}

	json.NewEncoder(w).Encode(response)
}
