package pictureManager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	destinationFolder = "../../storage/jpg"
	ErrNotFound       = fmt.Errorf("file not found")
)

func SavePicture(productID int, url string) error {

	if !strings.HasSuffix(strings.ToLower(url), ".jpg") && !strings.HasSuffix(strings.ToLower(url), ".jpeg") {
		// Если нет расширения .jpg/.jpeg, пробуем определить из Content-Type или добавляем .jpg
		return fmt.Errorf("URL does not point to a valid .jpg or .jpeg image: %s", url)
	}

	id := strconv.Itoa(productID)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexepted status code: %s", resp.Status)
	}

	fileName := fmt.Sprintf("%s.jpg", id)

	err = os.MkdirAll(destinationFolder, os.ModePerm) // 0755 permissions
	if err != nil {
		return fmt.Errorf("ошибка при создании папки '%s': %w", destinationFolder, err)
	}

	// Полный путь к файлу
	filePath := filepath.Join(destinationFolder, fileName)

	// 5. Создаем локальный файл
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла '%s': %w", filePath, err)
	}
	defer out.Close() // Закрываем файл при завершении функции

	// 6. Копируем данные из тела ответа в локальный файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при записи данных в файл '%s': %w", filePath, err)
	}

	fmt.Printf("Изображение успешно загружено и сохранено как '%s'\n", filePath)
	return nil

}

func Picture(productID string) ([]byte, error) {

	fileName := fmt.Sprintf("%s.jpg", productID)

	filePath := filepath.Join(destinationFolder, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, filePath)
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s, error: %w", filePath, err)
	}

	fmt.Printf("Picture found: %s\n", filePath)
	return file, nil
}
