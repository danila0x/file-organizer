package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStats struct {
	CountFiles    int
	TotalFileSize int64
}

func (fs *FileStats) String() string {
	var sizeStr string
	size := fs.TotalFileSize
	switch {
	case size < 1024:
		sizeStr = fmt.Sprintf("%d B", size)
	case size < 1024*1024:
		sizeStr = fmt.Sprintf("%.2f KB", float64(size)/1024)
	default:
		sizeStr = fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("\tФайлов: %d, Размер: %s", fs.CountFiles, sizeStr)
}

type FileOrganizer struct {
	sourceDir      string
	rulesMap       map[string]string
	processedFiles int
	logFile        *os.File
	statistics     map[string]*FileStats
	totalSize      int64
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
		statistics:     make(map[string]*FileStats),
		totalSize:      0,
	}, nil
}

func main() {
	fmt.Println("===File organizer===")
	fmt.Print("Введите путь к директории для организации (Enter для текущей директории):")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	sourcePath := strings.TrimSpace(input)
	if sourcePath == "" {
		var err error
		sourcePath, err = os.Getwd()
		if err != nil {
			fmt.Println("Ошибка получения текущей директории:", err)
			return
		}
		fmt.Printf("Используется текущая директория: %s\n", sourcePath)

	}
	org, err := NewFileOrganizer(sourcePath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer org.Close()

	org.rulesMap = map[string]string{
		".jpg": "Images", ".jpeg": "Images", ".png": "Images",
		".pdf": "Documents", ".doc": "Documents", ".docx": "Documents", ".txt": "Documents",
		".mp3": "Music", ".wav": "Music",
		".mp4": "Video", ".avi": "Video",
		".zip": "Archives", ".rar": "Archives",
	}

	if err := org.Organize(); err != nil {
		fmt.Println("Ошибка:", err)
	}

	fmt.Println(org.generateReport())
	fmt.Println("Организация завершена! Подробности в файле organizer.log")
}

func (fo *FileOrganizer) initLog() error {
	logPath := "organizer.log"
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

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string, fileSize int64) error {
	fmt.Printf("sourcePath: %s, targetDir: %s\n", sourcePath, targetDir)
	fileName := filepath.Base(sourcePath)
	fullPath := filepath.Join(fo.sourceDir, targetDir, fileName)
	fmt.Printf("filename: %s, fullpath: %s\n", fileName, fullPath)
	filePath := filepath.Join(fo.sourceDir, targetDir)
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		fo.logError("Ошибка создания папки")
		return fmt.Errorf("Ошибка создания папки")
	}
	fmt.Println("Папка успешно создана")

	newFilePath := filepath.Join(filePath, fileName)
	if _, err := os.Stat(newFilePath); err == nil {
		ext := filepath.Ext(fileName)
		base := strings.TrimSuffix(fileName, ext)
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		newFileName := base + "_" + timestamp + ext
		targetPath := filepath.Join(filePath, newFileName)
		if renameErr := os.Rename(sourcePath, targetPath); renameErr != nil {
			fo.logError(fmt.Sprintf("Не удалось переместить файл %s: %v", fileName, renameErr))
			return fmt.Errorf("ошибка перемещения файла: %w", renameErr)
		}

		if _, exists := fo.statistics[targetDir]; !exists {
			fo.statistics[targetDir] = &FileStats{}
		}
		fo.statistics[targetDir].CountFiles++
		fo.statistics[targetDir].TotalFileSize += fileSize
		fo.totalSize += fileSize
		fo.processedFiles++

		fo.logSuccess(fmt.Sprintf("Файл %q перемещён в %q (переименован в %q)", fileName, targetDir, newFileName))

	} else if os.IsNotExist(err) {
		if renameErr := os.Rename(sourcePath, newFilePath); renameErr != nil {
			fo.logError(fmt.Sprintf("Не удалось переместить файл %s: %v", fileName, renameErr))
			return fmt.Errorf("ошибка перемещения файла: %w", renameErr)
		}

		if _, exists := fo.statistics[targetDir]; !exists {
			fo.statistics[targetDir] = &FileStats{}
		}
		fo.statistics[targetDir].CountFiles++
		fo.statistics[targetDir].TotalFileSize += fileSize
		fo.totalSize += fileSize
		fo.processedFiles++

		fo.logSuccess(fmt.Sprintf("Файл %q перемещён в %q", fileName, targetDir))
	} else {
		fo.logError(fmt.Sprintf("Ошибка при проверке файла %s: %v", fileName, err))
		return fmt.Errorf("ошибка проверки файла: %w", err)
	}
	return nil
}

func (fo *FileOrganizer) Organize() error {
	if err := fo.initLog(); err != nil {
		return fmt.Errorf("ошибка инициализации лога: %w", err)
	}
	err := filepath.WalkDir(fo.sourceDir, func(path string, d fs.DirEntry, err error) error {
		fileInfo, err := d.Info()
		if err != nil {
			// Не можем получить размер
			fo.logError(fmt.Sprintf("Не удалось получить информацию о файле %s: %v", path, err))
			return nil
		}
		fileSize := fileInfo.Size()
		if filepath.Dir(path) != fo.sourceDir {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "organizer.log" {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		targetDir, exists := fo.rulesMap[ext]
		if !exists {
			return nil
		}
		if moveErr := fo.moveFile(path, targetDir, fileSize); moveErr != nil {
			fo.logError(fmt.Sprintf("Не удалось переместить %s: %v", path, moveErr))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ошибка при обходе директории: %w", err)
	}

	return nil
}

func (fo *FileOrganizer) generateReport() string {
	fmt.Println("\n=== Отчёт о перемещении файлов  ===")
	fmt.Println()
	totalStr := fmt.Sprintf("Всего обработано файлов: %d\n", fo.processedFiles)
	totalSize := fmt.Sprintf("Общий размер: %s\n\n", formatSize(fo.totalSize))
	categoryStr := "Статистика по категориям:\n\n"
	var finalStr string
	for category, stats := range fo.statistics {
		finalStr += fmt.Sprintf("%s:\n %s\n", category, stats.String())
	}
	return totalStr + totalSize + categoryStr + finalStr
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes < KB:
		return fmt.Sprintf("%d B", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	case bytes < GB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	default:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	}
}
