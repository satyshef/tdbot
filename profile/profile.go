//Package profile профиль телеграм бота tdbot
package profile

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

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

// TODO: Навести порядок с параметром CheckLimits
// Открыть существующий профиль из указанной директории
// @dir - путь к профилю
// @mode - режим работы. 0 - использовать режим по умолчанию, 1 - лимиты не проверять, 2 - лимиты проверять
func Get(dir string, limitsMode config.LimitsMode) (*Profile, error) {
	AddTail(&dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", dir)
	}
	if !IsProfile(dir) {
		return nil, fmt.Errorf("%s does not profile", dir)
	}

	conf := config.New()
	paths := GetConfigFiles(dir)
	err := conf.Load(paths...)
	if err != nil {
		return nil, err
	}
	phoneNumber, err := GetPhoneNumberFromPath(dir)
	if err != nil {
		return nil, err
	}
	prof, err := new(phoneNumber, dir, conf)
	if err != nil {
		return nil, err
	}
	//Сохраняем конфигурацию для того чтобы зафиксировать время последнего доступа к профилю
	//prof.SaveConfig()

	//блокируем профиль(Заглушка)
	if err := prof.lock(); err != nil {
		return nil, err
	}

	// устанавливаем режим проверки лимитов. Если режим не указан тогда ипользуем установленный в профиле
	switch limitsMode {
	case config.LimitsModeDontCheckLimits:
		prof.Config.APP.CheckLimits = false
	case config.LimitsModeCheckLimits:
		prof.Config.APP.CheckLimits = true
	}

	fmt.Println("Check limits", prof.Config.APP.CheckLimits)

	if prof.Config.APP.CheckLimits {
		//если у профиля ограничения по лимитам тогда игнорируем его
		exLimits := prof.CheckAllLimits()
		if exLimits != nil {
			prof.Close(1000)
			return nil, fmt.Errorf("Limit is exceeded : \n%#v\n", exLimits)
		}
	}

	return prof, nil
}

// IsProfile проверяем является ли указанная директория профилем. Профилем является директория с файлом конфигурации
func IsProfile(path string) bool {
	AddTail(&path)
	profileFile := GetMainConfigFile(path)
	if _, err := os.Stat(profileFile); err != nil {
		return false
	}
	return true
}

func GetMainConfigFile(path string) string {
	return filepath.Clean(path) + "/" + ProFile

}

// Ищем файлы с конфигурацией. Делаем поиск по всему пути
func GetConfigFiles(path string) []string {

	var result []string
	result = append(result, GetMainConfigFile(path))

	for path != "" && path != "/" {
		path = filepath.Clean(path)
		path = strings.TrimRight(path, filepath.Base(path))
		files, _ := filepath.Glob(path + "/bot_*.toml")
		result = append(result, files...)
	}
	if path == "" {
		path = "./"
	}
	files, _ := filepath.Glob(path + "bot_*.toml")
	result = append(result, files...)
	return result
}

// Получить список профилей в директории
func GetList(dir string, srt Sort) (result []string) {
	AddTail(&dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return result
		//log.Fatal(err)
	}

	switch srt {
	case SORT_RANDOM:
		for _, file := range files {
			if IsProfile(dir + file.Name()) {
				result = append(result, file.Name())
			}
		}
		result = shuffleArray(result)
	//сортеруем по времени использования профиля, на убывание
	case SORT_TIME_DESC:
		tmp_list := map[int64]string{}
		for _, file := range files {
			if IsProfile(dir + file.Name()) {
				prof, err := os.Stat(dir + file.Name() + "/" + ProFile)
				if err != nil {
					continue
				}
				tim := prof.Sys().(*syscall.Stat_t).Atim.Nano()
				tmp_list[tim] = file.Name()
			}
		}
		result = sortProfileTimeDESC(tmp_list)
	//сортеруем по времени использования профиля, по возрастанию
	case SORT_TIME_ASC:
		tmp_list := map[int64]string{}
		for _, file := range files {
			if IsProfile(dir + file.Name()) {
				prof, err := os.Stat(dir + file.Name() + "/" + ProFile)
				if err != nil {
					continue
				}
				//tim:=prof.ModTime().Unix()
				tim := prof.Sys().(*syscall.Stat_t).Atim.Nano()
				tmp_list[tim] = file.Name()
			}
		}
		result = sortProfileTimeASC(tmp_list)
	// по умолчанию сортируем в алфавитном порядке
	default:
		for _, file := range files {
			if IsProfile(dir + file.Name()) {
				result = append(result, file.Name())
			}
		}
	}
	return
}

func sortProfileTimeDESC(m map[int64]string) (result []string) {

	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	//sort.Sort(keys)
	sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
	for _, k := range keys {
		result = append(result, m[k])
	}
	return
}

func sortProfileTimeASC(m map[int64]string) (result []string) {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		result = append(result, m[k])
	}
	return
}

/*
// сортировка файлов профилей по времени последнего доступа
func sortByTime(lst []string) []string {
	if len(lst) == 0 {
		return nil
	}

	var result map[int64]

	for _, p := range lst {
		AddTail(&p)
		file, err := os.Stat(p + ProFile)
		if err != nil {
			continue
		}


	}

}
*/

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
