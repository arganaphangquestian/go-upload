package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/segmentio/ksuid"
)

func handleUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Can't Get uploaded file")
		returnError(w, "Can't Get uploaded file", 500)
		return
	}
	defer file.Close()

	filenameSplit := strings.Split(handler.Filename, ".")
	fileExtension := filenameSplit[len(filenameSplit)-1]

	// Store file
	root, err := os.Getwd()
	if err != nil {
		log.Println("Can't Get current project directory")
		returnError(w, "Can't Get current project directory", 500)
		return
	}
	filename := fmt.Sprintf("%s.%s", ksuid.New().String(), fileExtension)
	targetFile, err := os.Create(fmt.Sprintf("%s/%s/%s", root, "storage", filename))
	if err != nil {
		log.Println("Target file can't be create")
		returnError(w, "Target file can't be create", 500)
		return
	}
	defer targetFile.Close() // Decode Image
	var img image.Image
	if strings.ToLower(fileExtension) == "jpg" || strings.ToLower(fileExtension) == "jpeg" {
		img, err = jpeg.Decode(file)
		if err != nil {
			log.Println("Can't compress this image")
			returnError(w, "Can't compress this image", 500)
			return
		}
		m := resize.Resize(1000, 0, img, resize.Lanczos3)
		jpeg.Encode(targetFile, m, nil)
	} else if strings.ToLower(fileExtension) == "png" {
		img, err = png.Decode(file)
		if err != nil {
			log.Println("Can't compress this image")
			returnError(w, "Can't compress this image", 500)
			return
		}
		m := resize.Resize(1000, 0, img, resize.Lanczos3)
		jpeg.Encode(targetFile, m, nil)
	} else {
		// Others file
		if _, err := io.Copy(targetFile, file); err != nil {
			log.Println("Can't Store file into server directory")
			returnError(w, "Can't Store file into server directory", 500)
			return
		}
	}
	// Store Uploaded File into storage directory
	returnSuccess(w, "File Uploaded", 201, fmt.Sprintf("/static/%s", filename))
}

func handleList(w http.ResponseWriter, r *http.Request) {
	log.Println("Get List of images 🌆")
	files, err := ioutil.ReadDir("./storage")
	if err != nil {
		log.Println("Something went wrong!")
		returnError(w, "Something went wrong!", 500)
		return
	}
	var filesData []string
	for _, f := range files {
		filesData = append(filesData, fmt.Sprintf("/static/%s", f.Name()))
	}
	returnSuccess(w, "List get successfully", 200, filesData)
}

func returnError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "failed",
		"message": msg,
		"data":    nil,
	})
}

func returnSuccess(w http.ResponseWriter, msg string, statusCode int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": msg,
		"data":    data,
	})
}

func main() {
	mux := mux.NewRouter()
	// Static Files
	mux.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./storage"))))

	// API Handler
	mux.HandleFunc("/upload", handleUpload).Methods("POST")
	mux.HandleFunc("/list", handleList).Methods("GET")

	fmt.Println("Running at http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
