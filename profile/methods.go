// Package profile вспомогательные методы
package profile

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/polihoster/tdbot/fslock"
	"github.com/polihoster/tdbot/user"
)

var (
	lockFile string = "lock"
	locker   *fslock.Lock
)

//создать директорию профиля
func MakeProfileDir(path string) {
	//проверяем наличие директории профайла, если ее нет то пытаемся создать
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

/*
// LockProfile .... Не закрывается дескриптор при повторной инициализации!!!
func LockProfile(path string) error {
	AddTail(&path)
	locker = fslock.New(path + lockFile)
	err := locker.TryLock()
	if err != nil {
		return err
	}
	return nil
}

// UnlockProfile ...
func UnlockProfile(path string) error {
	if locker == nil {
		return nil
	}
	err := locker.Unlock()
	if err != nil {
		return err
	}
	//locker = nil
	return nil
}

*/
// Получить профиль
func Get(dir string, usrType user.Type) (*Profile, error) {

	AddTail(&dir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", dir)
	}

	phoneNumber := getPhoneNumber(dir)
	prof, err := New(dir, phoneNumber, phoneNumber, dir+ProFile, usrType)
	if err != nil {
		return nil, err
	}

	return prof, nil
}

// Получить список профилей в директории
func GetList(dir string, random bool) (result []string) {

	AddTail(&dir)
	/*
		//определяем является ли директория профилем
		if IsProfile(dir) {
			return []string{getPhoneNumber(dir)}
		}
	*/
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {

		if IsProfile(dir + file.Name()) {
			result = append(result, file.Name())
		}
		/*
			if file.IsDir() && isDigits(file.Name()) {
				result = append(result, file.Name())
			}
		*/
	}

	if random {
		return shuffleArray(result)
	}

	return result
}

// IsProfile проверяем является ли указанная директория профилем. Профилем являетяс директория с файлом
func IsProfile(path string) bool {
	AddTail(&path)
	/*
		if _, err := os.Stat(path); err != nil {
			return false
		}
	*/

	if _, err := os.Stat(path + ProFile); err != nil {
		return false
	}

	//if os.IsNotExist(err) {
	//		return false
	//	}
	return true
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

//извлекаем номер телефона из пути к профилю
func getPhoneNumber(path string) (result string) {

	//проверяем является ли последний элемент пути числом
	parts := strings.Split(path, "/")
	result = parts[len(parts)-2]
	if !isDigits(result) {
		return ""
	}

	return result
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
