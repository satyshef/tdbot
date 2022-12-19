// Package profile вспомогательные методы
package profile

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/satyshef/tdbot/config"
)

//Переместить профиль
// @dst - директория назначения
func (p *Profile) Move(dst string) error {
	AddTail(&dst)
	makeProfileDir(dst)
	dst += path.Base(p.Location())
	fmt.Printf("Move %s to %s\n", p.Location(), dst)
	err := p.unlock()
	if err != nil {
		return err
	}
	//dst = "/home/boba/go/src/radic/tebot/cmd/app/profiles/logout/dssss"
	err = os.Rename(p.Location(), dst)
	if err != nil {
		return err
	}
	p.SetLocation(dst)
	return p.Reload()
}

// Remove Удалить профиль
func (p *Profile) Remove() error {
	fmt.Printf("Remove profile %s\n", p.Location())
	return os.RemoveAll(p.Location())
}

// @timeout - время ожидания закрытия в миллисекудах, 0 - не ждать
func (p *Profile) Close(timeout int) error {
	//fmt.Println("CLOSE PROFILE")
	//Тест. Пробуем блокировку Level DB
	//UnlockProfile(p.Location())

	p.unlock()
	//TODO:		Тест
	p.Event.Store.Close()
	//p = nil
	//p.User = nil
	//p.Event = nil
	//p.Config = nil
	if timeout != 0 {
		time.Sleep(time.Millisecond * time.Duration(timeout))
	}

	return nil
}

// Перезагрузить профиль
// @path - если указан, тогда профиль загружается по данному пути
func (p *Profile) Reload() error {
	p.Event.Store.Close()
	newProf, err := Get(p.Location(), config.LimitsModeDefault)
	if err != nil {
		return err
	}
	p.dir = newProf.dir
	p.Config = newProf.Config
	p.Event = newProf.Event
	p.User = newProf.User
	return nil
}

// Директория профиля
func (p *Profile) Location() string {
	return p.dir
}

// Установить директорию профиля
// @path - путь к директории профиля
func (p *Profile) SetLocation(path string) {
	AddTail(&path)
	p.dir = path
}

// Путь к файлу конфигурации профиля
func (p *Profile) ConfigFile() string {
	return p.Location() + ProFile
}

//LoadConfig загрузить конфигурацию из файла
func (p *Profile) LoadConfig() error {
	if p.ConfigFile() == "" {
		return fmt.Errorf("%s", "Config file not set")
	}
	err := p.Config.Load(p.ConfigFile())
	if err != nil {
		return fmt.Errorf("Error loading config file %s : %s", p.ConfigFile(), err)
	}
	fmt.Printf("Profile config : %s\n", p.ConfigFile())
	return nil
}

func (p *Profile) SaveConfig() error {
	if p.ConfigFile() == "" {
		return fmt.Errorf("%s", "Config file not set")
	}
	if err := p.Config.Save(p.ConfigFile()); err != nil {
		return err
	}
	return nil
}

func (p *Profile) lock() error {
	return nil
}

func (p *Profile) unlock() error {
	if p.Event != nil {
		return p.Event.Close()
	}
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

// CheckLimit проверить достижение литов события
// @evType - тип события
// @evName - имя события лимит которого нужно проверить
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
		}
		//Проверяем лимит, если превышен то добавляем его в список результатов
		if len(evns) >= int(limit.Limit) {
			// в ответ записываем время до окончания интервала лимита
			limit.Interval -= int32((time.Now().UnixNano() - evns[len(evns)-1].Time) / int64(time.Second))
			result = append(result, limit)
		}
	}
	return result
}

// CheckDir проверяем наличие директории профиля
func (p *Profile) CheckDir() bool {
	if _, err := os.Stat(p.Location()); err == nil {
		return true
	}
	return false
}

// BaseDir базавая директория профиля
func (p *Profile) BaseDir() string {
	base := strings.Replace(p.Location(), p.User.PhoneNumber+"/", "", 1)
	if base == "" {
		base = "./"
	}
	return base
}

//============================= ПЕРЕСМОТРЕТЬ ==========================================
/*
// Find инициализировать профиль бота
func Find1(profileDir, configFile string) (prof *Profile, err error) {
	AddTail(&profileDir)
	phoneNumber, err := GetPhoneNumberFromPath(profileDir)
	// Если указан номер телефона тогда инициализируем конкретный профиль, иначе подбираем из списка
	if err == nil {
		prof, err = New(profileDir, phoneNumber, phoneNumber, configFile)
		if err != nil {
			return nil, err
		}
	} else {
		if prof = Load(profileDir, configFile); prof == nil {
			time.Sleep(3 * time.Second)
			prof, err = Find1(profileDir, configFile)
			if err != nil {
				return nil, err
			}
		}
	}
	return prof, nil
}

*/

/*
// Load поиск и загрузка профиля из директории
func Load(dir, configFile string) (prof *Profile) {

	var err error
	//ищем профиль в указанной директории
	profileList := GetList(dir, true)
	//profileList = shuffleArray(profileList)

	for _, p := range profileList {
		dir := dir + p
		prof, err = New(dir, p, p, configFile)
		if err != nil {
			continue
		}
		break
	}

	return prof
}

*/
