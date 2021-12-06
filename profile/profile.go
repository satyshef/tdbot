//Package profile профиль телеграм бота tdbot
package profile

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"time"

	"github.com/polihoster/tdbot/config"
	"github.com/polihoster/tdbot/events/eventman"
	"github.com/polihoster/tdbot/user"
)

//Profile ...
type Profile struct {
	User       *user.User
	Dir        string
	ConfigFile string
	Config     *config.Config
	Event      *eventman.Manager
}

const (
	EventTypeRequest  = "request"
	EventTypeResponse = "response"
	EventTypeError    = "error"
)

const (
	//коды ошибок данного модуля начинаются с 2
	ErrorCodeLimitExceeded = 201
	ErrorCodeLimitNotSet   = 202
	ErrorCodeNotInit       = 203
	ErrorCodeDirNotExists  = 204
	ErrorCodeSystem        = 205
)

//Директория с конфигурацией по умолчанию
var defaultDir = ""

//var DefaultConfigFile = "bot.toml"
var ProFile = "bot.toml"

//перенести в конфигурацию
var (
	watchEventList = []string{"*"}

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

// =================================== OLD METHODS =========================

func newProfile(dir, name, phone string, usrType user.Type) *Profile {

	cfg := config.New()
	return &Profile{
		User:       user.New(name, phone, usrType),
		Dir:        dir,
		ConfigFile: ProFile,
		Config:     cfg,
		Event:      newEventManager(dir, watchEventList),
	}
}

func newEventManager(dir string, eventList []string) *eventman.Manager {
	return eventman.New(dir+"event", eventList)
}

// Find инициализировать профиль бота
func Find(profileDir, configFile string, usrType user.Type) (prof *Profile, err error) {

	AddTail(&profileDir)

	phoneNumber := getPhoneNumber(profileDir)

	// Если указан номер телефона тогда инициализируем конкретный профиль, иначе подбираем готовый к работе из списка
	if phoneNumber != "" {

		// Modify
		//MakeProfileDir(profileDir)

		prof, err = New(profileDir, phoneNumber, phoneNumber, configFile, usrType)
		if err != nil {
			return nil, err
		}

	} else {
		if prof = Load(profileDir, configFile, usrType); prof == nil {
			//fmt.Println("Нет доступных профилей. Ожидаем....")
			time.Sleep(3 * time.Second)
			prof, err = Find(profileDir, configFile, usrType)
			if err != nil {

				return nil, err
			}
		}

	}

	return prof, nil
}

// New ...
//func New(dir, name, phone string, usrType user.Type, copyToProfile bool) (prof *Profile, err error) {
func New(dir, name, phone string, configFile string, usrType user.Type) (prof *Profile, err error) {

	AddTail(&dir)
	MakeProfileDir(dir)
	//блокируем профиль для избежания повторного использования
	err = LockProfileDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Profile is already in use")
	}

	fmt.Printf("Loading profile %s\n", phone)

	prof = newProfile(dir, name, phone, usrType)
	err = prof.LoadConfig(configFile)
	if err != nil {
		//fmt.Printf("%s\n", err)
		return nil, err
	}

	if prof.ConfigFile != prof.Dir+ProFile {
		prof.SaveConfig(prof.Dir + ProFile)
	}
	//если у профиля ограничения по лимитам тогда игнорируем его
	if prof.Config.APP.Mode == 2 {
		exLimits := prof.CheckAllLimits()
		if exLimits != nil {
			prof.Close()
			return nil, fmt.Errorf("Limit is exceeded : \n%#v\n", exLimits)
		}

	}

	return prof, err
}

// Load поиск и загрузка профиля из директории
func Load(dir, configFile string, usrType user.Type) (prof *Profile) {
	var err error
	//ищем профиль в указанной директории
	profileList := GetList(dir, true)
	//profileList = shuffleArray(profileList)

	for _, p := range profileList {
		dir := dir + p
		prof, err = New(dir, p, p, configFile, usrType)
		if err != nil {
			continue
		}
		break
	}

	return prof
}

//LoadConfig найти и загрузить конфигурацию из файла
// @configFile - Путь к файлу конфигурации. Если не указан происходит поиск в дефолтной директории и в директории профиля
func (p *Profile) LoadConfig(configFile string) error {

	//var configPath string
	var err error

	//Если указан путь к файлу конфигурации то загружаем от туда, иначе ищем в дефолтной директории и директрии профиля
	/*
		if configFile == "" {
			configFile, err = p.Config.Find(p.ConfigFile, p.Dir, defaultDir)
		} else {
			err = p.Config.Load(configFile)
		}
	*/

	if configFile == "" {
		configFile, err = p.Config.Find(ProFile, p.Dir, defaultDir)
		if err != nil {
			return err
		}
	}

	err = p.Config.Load(configFile)

	if err != nil {
		return fmt.Errorf("Error loading config file %s : %s", configFile, err)
	}

	PrepareConfig(p.Config)

	fmt.Printf("Profile config : %s\n", configFile)
	p.ConfigFile = configFile
	return nil
}

func (p *Profile) SaveConfig(configFile string) error {

	if err := p.Config.Save(configFile); err != nil {
		return err
	}

	p.ConfigFile = configFile
	return nil
}

// CheckAllLimits проверить все лимиты
func (p *Profile) CheckAllLimits() config.Limits {
	result := make(map[string]map[string][]config.Limit)

	for evType, eventList := range p.Config.Limits {
		for evName := range eventList {
			if exLimits := p.CheckLimit(evType, evName); exLimits != nil {
				result[evType] = make(map[string][]config.Limit)
				result[evType][evName] = exLimits
			}

		}

	}

	if len(result) > 0 {
		return result
	}

	return nil

}

//Переместить профиль
func (p *Profile) Move(dst string) error {

	err := UnlockProfileDir(p.Dir)
	if err != nil {
		return err
	}
	AddTail(&dst)
	MakeProfileDir(dst)
	dst += path.Base(p.Dir)
	fmt.Printf("Move %s to %s\n", p.Dir, dst)
	err = os.Rename(p.Dir, dst)
	p = Load(dst, p.ConfigFile, p.User.Type)
	return err
}

// Remove Удалить профиль
func (p *Profile) Remove() error {
	/*
		err := UnlockProfileDir(p.Dir)
		if err != nil {
			return err
		}
	*/

	fmt.Printf("Remove profile %s\n", p.Dir)
	//err = os.RemoveAll(p.Dir)
	return os.RemoveAll(p.Dir)
}

func (p *Profile) Close() error {
	if p.Event != nil {
		p.Event.Store.Close()
	}
	UnlockProfileDir(p.Dir)
	p.User = nil
	p.Event = nil
	p.Config = nil
	return nil
}

// CheckLimit проверить достижение литов события
func (p *Profile) CheckLimit(evType, evName string) (result []config.Limit) {

	//Если для события не установлен лимит тогда проверку считаем удачной
	if p.Config == nil || p.Config.Limits == nil || p.Config.Limits[evType] == nil || p.Config.Limits[evType][evName] == nil {
		return nil
	}

	for _, limit := range p.Config.Limits[evType][evName] {
		minTime := time.Now().Unix() - int64(limit.Interval)
		maxTime := time.Now().Unix() + 1
		evns, err := p.Event.SearchByTime(evType, evName, minTime, maxTime)
		if err != nil {
			log.Fatalf("Search Event Error : %s\n", err)
			os.Exit(1)
			//continue
			//return 0, err
		}

		if len(evns) >= int(limit.Limit) {
			// в ответ записываем время до ближайшего события
			limit.Interval -= int32((time.Now().UnixNano() - evns[len(evns)-1].Time) / int64(time.Second))
			result = append(result, limit)
		}
	}

	return result
}

// CheckDir проверяем наличие директории
func (p *Profile) CheckDir() bool {

	if _, err := os.Stat(p.Dir); err == nil {
		return true
	}
	//if os.IsNotExist(err) {
	//		return false
	//	}
	return false
}

// BaseDir базавая директория профиля
func (p *Profile) BaseDir() string {
	base := strings.Replace(p.Dir, p.User.PhoneNumber+"/", "", 1)
	if base == "" {
		base = "./"
	}
	return base
}

func (p *Profile) ConfigPath() string {
	return p.Dir + p.ConfigFile
}

func PrepareConfig(conf *config.Config) {

	//set first name
	if _, err := os.Stat(conf.APP.FirstName); err == nil {
		conf.APP.FirstName = loadRandomString(conf.APP.FirstName)
		fmt.Println("Auto First Name : ", conf.APP.FirstName)
	}

	//set system version
	if _, err := os.Stat(conf.APP.SystemVersion); err == nil {
		conf.APP.SystemVersion = loadRandomString(conf.APP.SystemVersion)
		fmt.Println("Auto System Version : ", conf.APP.SystemVersion)
	}

	//set app version
	if _, err := os.Stat(conf.APP.AppVersion); err == nil {
		conf.APP.AppVersion = loadRandomString(conf.APP.AppVersion)
		fmt.Println("Auto App Version : ", conf.APP.AppVersion)
	}

	// Set lang pack
	if _, err := os.Stat(conf.APP.SystemLanguageCode); err == nil {
		conf.APP.SystemLanguageCode = loadRandomString(conf.APP.SystemLanguageCode)
		fmt.Println("Auto Lang : ", conf.APP.SystemLanguageCode)
	}

	//set device model
	if _, err := os.Stat(conf.APP.DeviceModel); err == nil {
		conf.APP.DeviceModel = loadRandomString(conf.APP.DeviceModel)
		fmt.Println("Auto Device : ", conf.APP.DeviceModel)
	}

	//set api id and hash
	if _, err := os.Stat(conf.APP.ID); err == nil {
		s := strings.Split(loadRandomString(conf.APP.ID), ":")
		conf.APP.ID = s[0]
		conf.APP.Hash = s[1]
		fmt.Println("Auto API ID : ", s)
	}

}

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
