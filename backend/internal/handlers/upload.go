package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	uploadDir := "./uploads"
	filename := filepath.Join(uploadDir, file.Filename)

	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go processImage(filename)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "File uploaded successfully",
		"filename": file.Filename,
		"url":      "/uploads/" + file.Filename,
	})
}

func processImage(filename string) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
		baseName := filepath.Base(filename)
		if strings.HasPrefix(baseName, "crash_") {
			log.Printf("Processing will crash for file: %s", filename)
			panic("image processing failed — unable to process " + filename)
		}

		outputFile := strings.TrimSuffix(filename, ext) + "_thumb" + ext
		cmd := exec.Command("sh", "-c", fmt.Sprintf("convert %s -resize 300x300 %s", filename, outputFile))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Image processing failed for %s: %v, output: %s", filename, err, string(output))
		} else {
			log.Printf("Thumbnail created: %s", outputFile)
		}
	}
}

func ensureUploadDir() {
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
}
