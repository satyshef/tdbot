//Package profile профиль телеграм бота tdbot
package profile

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/satyshef/tdbot/config"
	"github.com/satyshef/tdbot/events/eventman"
	"github.com/satyshef/tdbot/user"
)

//Директория с конфигурацией по умолчанию
//var defaultDir = ""
/*
var (
	lockFile string = "lock"
	locker   *fslock.Lock
)
*/

//Имя файла конфигурации профиля
var ProFile = "bot.toml"

//перенести в конфигурацию
var (
//watchEventList = []string{"*"}

/*
	watchEventList = []string{
		"importContacts",
		"searchContacts",
		"sendMessage",
		"setName",
		"getUser",
		"createPrivateChat",
		"sendToNewUser",
		"createPrivateChat",
		"setAuthenticationPhoneNumber",
	}
*/
)

func init() {
	//Проверка совместимости с текущей ОС
	if runtime.GOOS != "linux" {
		log.Fatalf("Profile : OS %s not supported", runtime.GOOS)
	}
}

// Создать новый экземпляр profile
func new(phone, profileDir string, conf *config.Config) (*Profile, error) {
	AddTail(&profileDir)
	eventManager, err := newEventManager(profileDir, conf.Watchlist)
	if err != nil {
		return nil, err
	}
	prof := &Profile{
		User: user.New(conf.APP.FirstName, phone, user.TypeTelegram),
		//dir:  profileDir,
		//configFile: ProFile,
		Config: conf,
		Event:  eventManager,
	}
	prof.SetLocation(profileDir)
	return prof, nil
}

// Создать новый экземпляр eventman.Manager
func newEventManager(dir string, watchList []string) (*eventman.Manager, error) {
	return eventman.New(dir+"event", watchList)
}

// Создать новый профиль
// @phone - номер телефона
// @profileDir - директория хранения профиля
// @conf - конфигурация
func New(phone, profileDir string, conf *config.Config) (prof *Profile, err error) {
	fmt.Printf("Creating new profile %s\n", phone)
	AddTail(&profileDir)
	profileDir += phone
	AddTail(&profileDir)
	err = makeProfileDir(profileDir)
	if err != nil {
		return nil, err
	}
	/*
		//блокируем профиль для избежания повторного использования
		//Заглушка. Используем блокировку Level DB
		err = LockProfile(profileDir)
		if err != nil {
			return nil, err
		}
	*/
	prof, err = new(phone, profileDir, conf)
	if err != nil {
		return nil, err
	}
	//блокируем профиль(Заглушка)
	if err := prof.lock(); err != nil {
		return nil, err
	}
	prof.SaveConfig()
	return prof, err
}

// Открыть существующий профиль из указанной директории
// @dir - путь к профилю
// @mode - режим работы. 0 - взять из файла конфигурации, 1 - лимиты не проверять, 2 - лимиты проверять
func Get(dir string, mode config.Mode) (*Profile, error) {
	AddTail(&dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", dir)
	}
	if !IsProfile(dir) {
		return nil, fmt.Errorf("%s does not profile", dir)
	}
	phoneNumber, err := GetPhoneNumberFromPath(dir)
	if err != nil {
		return nil, err
	}
	conf := config.New()
	conf.Load(dir + ProFile)
	prof, err := new(phoneNumber, dir, conf)
	if err != nil {
		return nil, err
	}
	//блокируем профиль(Заглушка)
	if err := prof.lock(); err != nil {
		return nil, err
	}
	// устанавливаем режим работы
	if mode == config.ModeDefault {
		mode = prof.Config.APP.Mode
	}
	if mode == config.ModeCheckLimits {
		//если у профиля ограничения по лимитам тогда игнорируем его
		exLimits := prof.CheckAllLimits()
		if exLimits != nil {
			prof.Close()
			return nil, fmt.Errorf("Limit is exceeded : \n%#v\n", exLimits)
		}
	}
	return prof, nil
}

// IsProfile проверяем является ли указанная директория профилем. Профилем является директория с файлом конфигурации
func IsProfile(path string) bool {
	AddTail(&path)
	if _, err := os.Stat(path + ProFile); err != nil {
		return false
	}
	return true
}

// Получить список профилей в директории
func GetList(dir string, random bool) (result []string) {
	AddTail(&dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return result
		//log.Fatal(err)
	}
	for _, file := range files {
		if IsProfile(dir + file.Name()) {
			result = append(result, file.Name())
		}
	}
	if random {
		return shuffleArray(result)
	}
	return result
}

//============================= ПЕРЕСМОТРЕТЬ ==========================================
/*
// LockProfile .... Не закрывается дескриптор при повторной инициализации!!!. Заглушка!!!
func LockProfile(path string) error {
	return nil

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
	return nil
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
