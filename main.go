package main

import (
	"fmt"
	"log"
	"os"
)

type FileOrganizer struct {
	sourceDir      string
	rulesMap       map[string]string
	processedFiles int
	logFile        *os.File
}

func NewFileOrganizer(sourceDir string) (*FileOrganizer, error) {
	if sourceDir == "" {
		return nil, fmt.Errorf("sourceDir is empty")
	}
	info, err := os.Stat(sourceDir)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return nil, err
	}
	if info.IsDir() == false {
		return nil, fmt.Errorf("its not a directory")
	}
	return &FileOrganizer{
		sourceDir:      sourceDir,
		rulesMap:       map[string]string{},
		processedFiles: 0,
		logFile:        nil,
	}, nil
}

func main() {
	DefaultRules := map[string]string{
		".jpg":  "Images",
		".jpeg": "Images",
		".png":  "Images",
		".pdf":  "Documents",
		".doc":  "Documents",
		".docx": "Documents",
		".txt":  "Documents",
		".mp3":  "Music",
		".wav":  "Music",
		".mp4":  "Video",
		".avi":  "Video",
		".zip":  "Archives",
		".rar":  "Archives",
	}

	for key, value := range DefaultRules {
		fmt.Println(key, value)
	}
}

func (fo *FileOrganizer) initLog() error {
	logFile, err := os.OpenFile(fo.sourceDir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Не удалось открыть файл лога %w", err)
	}
	fo.logFile = logFile
	log.SetOutput(logFile)
	return nil
}

func (fo *FileOrganizer) logSuccess(message string) {
	log.Printf("[SUCCESS] %s", message)
}

func (fo *FileOrganizer) logError(message string) {
	log.Printf("[ERROR] %s", message)
}

func (fo *FileOrganizer) Close() error {
	if fo.logFile != nil {
		err := fo.logFile.Close()
		if err != nil {
			return fmt.Errorf("Ошибка при закрытии лог-файла: %w", err)
		}
		fo.logFile = nil

		log.SetOutput(os.Stdout)
		log.Println("Лог-файл закрыт")
	}
	return nil
}
