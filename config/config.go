// Package config конфигурация телеграм бота
package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type (
	//Config конфигурация бота
	Config struct {
		APP       *APP     `toml:"app"`
		Limits    Limits   `toml:"limits"`
		Watchlist []string `toml:"watch_list"`
		Mimicry   *Mimicry `toml:"mimicry"`
		Log       *Log     `toml:"log"`
		Proxy     *Proxy   `toml:"proxy"`
	}

	//APP конфигурация клиента телеграм бота
	APP struct {
		ID                 string `toml:"id"`
		Hash               string `toml:"hash"`
		SystemVersion      string `toml:"sys_version"`
		AppVersion         string `toml:"app_version"`
		SystemLanguageCode string `toml:"lang_pack"`
		DeviceModel        string `toml:"device_model"`
		AuthPass           string `toml:"auth_pass"`
		HintPass           string `toml:"hint_pass"`
		FirstName          string `toml:"first_name"`
		Photo              string `toml:"photo"`
		ShowPhoneMode      int    `toml:"show_phone_mode"` // 0 - без изменений, 1 - показывать, 2 - скрывать
		Mode               Mode   `toml:"mode"`            // 1 - single mode (не проверять лимиты при старте) 2 - group mode(проверять)
		DontRebootInterval int32  `toml:"max_interval"`    // интервал при котором не происходит отключение профиля если сработал лимит
		//CheckLimit         bool   `toml:"check_limit"`
		SetOnline bool   `toml:"set_online"`
		DirBanned string `toml:"banned_dir"`
		DirLogout string `toml:"logout_dir"`
		DirFoul   string `toml:"foul_dir"`
		DirDouble string `toml:"double_dir"`
	}

	Mimicry struct {
		FriendRoom  string `toml:"friend_room"` // группа откуда берем друзей
		FriendName  string `toml:"friend_name"` //если в имени контакта будет данная фраза то контакт зачисляем к друзьям
		FriendCount int    `toml:"friend_count"`
		//Enable bool    `toml:"enable"`
		//Friend *Friend `toml:"friend"`
	}

	//Proxy ...
	Proxy struct {
		Host   string    `toml:"host"`
		Port   int32     `toml:"port"`
		User   string    `toml:"user"`
		Pass   string    `toml:"pass"`
		Type   ProxyType `toml:"type"`
		Enable bool      `toml:"enable"`
	}

	//Log ...
	Log struct {
		Level int    `toml:"level"`
		File  string `toml:"file"`
	}

	// Limits map. Key - event name
	Limits map[string]map[string][]Limit

	// Limit ...
	Limit struct {
		Interval int32
		Limit    int32
	}

	ProxyType string

	Mode int
)

// режим работы бота
const (
	ModeDefault         Mode = 0
	ModeDontCheckLimits Mode = 1
	ModeCheckLimits     Mode = 2
)

const (
	ProxyTypeSocks5  = "socks5"
	ProxyTypeHttp    = "http"
	ProxyTypeMtproto = "mtproto"
)

//New ...
func New() *Config {

	return &Config{
		APP: &APP{
			ID:                 "187786",
			Hash:               "e782045df67ba48e441ccb105da8fc85",
			SystemLanguageCode: "en",
			DeviceModel:        "Web Client",
			DirBanned:          "",
			DirLogout:          "",
			DirFoul:            "",
			DirDouble:          "",
			Mode:               1,
			DontRebootInterval: 10,
		},

		Log: &Log{
			Level: 4,
			File:  "/var/log/tbot.log",
		},

		//Humanity: &Humanity{Enable: false},
	}
}

func Load(path string) (*Config, error) {
	conf := New()
	err := conf.Load(path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// Find ищет конфигурацию в директориях и загружает если находит
// возвращает путь откуда была загружена конфигурация
// @fileName - имя файла конфигурации
// @confPath - список директорий где ищется файл конфигурации
func (c *Config) Find(fileName string, confPaths ...string) (string, error) {

	for _, path := range confPaths {
		path += fileName
		err := c.Load(path)
		if err == nil {
			//fmt.Printf("Конфигурация загружена из %s... \n", path+fileName)
			return path, nil
		}

		//fmt.Printf("Загрузка конфигурации из %s : %s\n", path, err)
	}

	return "", fmt.Errorf("Файл конфигурации %s не найден", fileName)
}

// Load загрузить конфигурацию из файла
func (c *Config) Load(path string) error {
	_, err := toml.DecodeFile(path, c)
	c.prepare()
	return err
}

//Save сохранить конфигурацию
// @fileName - путь к файлу в который сохраняем
func (c *Config) Save(fileName string) error {
	fmt.Printf("Save configuration to %s\n", fileName)
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	defer f.Close()
	/*
		//Делаем копию что бы не изменить основную структуру
		var b Config
		copier.Copy(&b, c)
		b.APP.ID = ""
		b.APP.Hash = ""
	*/
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("%s", err)
	}
	/*
		if err := f.Close(); err != nil {
			return fmt.Errorf("%s", err)
		}
	*/

	return nil
}

func (c *Config) prepare() {

	//c.APP.FirstName = strings.Trim(c.APP.FirstName, " \n\t")
	//c.APP.DeviceModel = strings.Trim(c.APP.DeviceModel, " \n\t")

	//set first name
	if _, err := os.Stat(c.APP.FirstName); err == nil {
		c.APP.FirstName = loadRandomString(c.APP.FirstName)
		fmt.Println("Auto First Name : ", c.APP.FirstName)
	}

	//set system version
	if _, err := os.Stat(c.APP.SystemVersion); err == nil {
		c.APP.SystemVersion = loadRandomString(c.APP.SystemVersion)
		fmt.Println("Auto System Version : ", c.APP.SystemVersion)
	}

	//set app version
	if _, err := os.Stat(c.APP.AppVersion); err == nil {
		c.APP.AppVersion = loadRandomString(c.APP.AppVersion)
		fmt.Println("Auto App Version : ", c.APP.AppVersion)
	}

	// Set lang pack
	if _, err := os.Stat(c.APP.SystemLanguageCode); err == nil {
		c.APP.SystemLanguageCode = loadRandomString(c.APP.SystemLanguageCode)
		fmt.Println("Auto Lang : ", c.APP.SystemLanguageCode)
	}

	//set device model
	if _, err := os.Stat(c.APP.DeviceModel); err == nil {
		c.APP.DeviceModel = loadRandomString(c.APP.DeviceModel)
		fmt.Println("Auto Device : ", c.APP.DeviceModel)
	}

	//set api id and hash
	if _, err := os.Stat(c.APP.ID); err == nil {
		s := strings.Split(loadRandomString(c.APP.ID), ":")
		c.APP.ID = s[0]
		c.APP.Hash = s[1]
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
		text := strings.Trim(scanner.Text(), " \n\t")
		lines = append(lines, text)
	}

	if scanner.Err() != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("Empty file %s", fileName)
	}

	return lines, nil
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

func (l Limits) JSON() string {
	//Limits map[string]map[string][]Limit
	r, _ := json.Marshal(l)
	return string(r)
}
