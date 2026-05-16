package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	org, err := NewFileOrganizer(".")
	if err != nil {
		fmt.Println(err)
	}
	defer org.Close()
	if err := org.initLog(); err != nil {
		fmt.Println("Ошибка инициализации лога:", err)
		return
	}

	// org.logSuccess("Файл \"report.pdf\" перемещён в директорию \"Documents\"")
	// org.logError("Невозможно переместить файл \"data.tmp\" - файл занят")
	// if err := org.moveFile(); err != nil {

	// }
}

func (fo *FileOrganizer) initLog() error {
	logPath := filepath.Join(fo.sourceDir, "organizer.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Не удалось открыть файл лога %w", err)
	}
	fo.logFile = logFile
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)
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

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string) error {
	fmt.Printf("sourcePath: %s, targetDir: %s", sourcePath, targetDir)
	fileName := filepath.Base(sourcePath)
	fullPath := filepath.Join(fo.sourceDir, targetDir, fileName)
	fmt.Printf("filename: %s, fullpath: %s", fileName, fullPath)
	filePath := filepath.Join(fo.sourceDir, targetDir)
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		//fmt.Println("Ошибка создания папки:", err)
		fo.logError("Ошибка создания папки")
	}
	fmt.Println("Папка успешно создана")
	newFilePath := filepath.Join(filePath, fileName)
	if _, err := os.Stat(newFilePath); err == nil {
		ext := filepath.Ext(fileName)
		base := strings.TrimSuffix(fileName, ext)
		timestamp := time.Now().Format("20060102_150405")
		newFileName := base + "_" + timestamp + ext
		targetPath := filepath.Join(filePath, newFileName)
		if renameErr := os.Rename(sourcePath, targetPath); renameErr != nil {
			fo.logError(fmt.Sprintf("Не удалось переместить файл %s: %v", fileName, renameErr))
			return fmt.Errorf("ошибка перемещения файла: %w", renameErr)
		}
		fo.logSuccess(fmt.Sprintf("Файл %q перемещён в %q (переименован в %q)", fileName, targetDir, newFileName))

	} else if os.IsNotExist(err) {
		if renameErr := os.Rename(sourcePath, newFilePath); renameErr != nil {
			fo.logError(fmt.Sprintf("Не удалось переместить файл %s: %v", fileName, renameErr))
			return fmt.Errorf("ошибка перемещения файла: %w", renameErr)
		}
		fo.logSuccess(fmt.Sprintf("Файл %q перемещён в %q", fileName, targetDir))
	} else {
		fo.logError(fmt.Sprintf("Ошибка при проверке файла %s: %v", fileName, err))
		return fmt.Errorf("ошибка проверки файла: %w", err)
	}
	return nil
}
