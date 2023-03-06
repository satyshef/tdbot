package profile

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func loadRandomString(fileName string) string {

	lines, err := readFileToSlice(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return shuffleArray(lines)[0]
}

func readFileToSlice(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("Empty file %s", fileName)
	}

	return lines, nil
}

//извлекаем номер телефона из пути к профилю
func GetPhoneNumberFromPath(path string) (string, error) {
	AddTail(&path)
	//проверяем является ли последний элемент пути числом
	parts := strings.Split(path, "/")
	result := parts[len(parts)-2]
	if !isDigits(result) {
		return "", fmt.Errorf("%s", "Phone number in path not found")
	}
	return result, nil
}

//AddTail добавить слеш в конец строки
func AddTail(in *string) {

	tm := *in
	if tm == "" {
		tm = "./"
	} else if tm[len(tm)-1:] != "/" {
		tm += "/"
	}
	*in = tm
}

func isDigits(s string) bool {

	_, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return false
	}

	return true

}

func shuffleArray(src []string) []string {
	final := make([]string, len(src))
	rand.Seed(time.Now().UTC().UnixNano())
	perm := rand.Perm(len(src))

	for i, v := range perm {
		final[v] = src[i]
	}
	return final
}

//создать директорию профиля
func makeProfileDir(path string) error {
	//проверяем наличие директории профайла, если ее нет то пытаемся создать
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	//Если небыло ошибки значит директория уже существует
	return fmt.Errorf("%s", "Profile exist")
}

func IsExistDir(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Printf("Check dir error: %s", err)
		return false
	}
	return true
}

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
