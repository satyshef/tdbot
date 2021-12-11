// Package config конфигурация телеграм бота
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type (
	//Config конфигурация бота
	Config struct {
		APP     *APP     `toml:"app"`
		Limits  Limits   `toml:"limits"`
		Mimicry *Mimicry `toml:"mimicry"`
		Log     *Log     `toml:"log"`
		Proxy   *Proxy   `toml:"proxy"`
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
		Mode               int    `toml:"mode"`            // 1 - single mode (не проверять лимиты при старте) 2 - group mode(проверять)
		DontRebootInterval int32  `toml:"max_interval"`    // интервал при котором не происходит отключение профиля если сработал лимит
		//CheckLimit         bool   `toml:"check_limit"`
		SetOnline bool   `toml:"set_online"`
		BannedDir string `toml:"banned_dir"`
		LogoutDir string `toml:"logout_dir"`
		FoulDir   string `toml:"foul_dir"`
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
			BannedDir:          "banned",
			LogoutDir:          "logout",
			FoulDir:            "foul",
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

	c.APP.FirstName = strings.Trim(c.APP.FirstName, " \n\t")
	c.APP.DeviceModel = strings.Trim(c.APP.DeviceModel, " \n\t")

	return err
}

//Save сохранить конфигурацию
// @fileName - путь к файлу в который сохраняем
func (c *Config) Save(fileName string) error {
	fmt.Printf("Сохраняем конфигурацию в %s\n", fileName)
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	defer f.Close()

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

func (l Limits) JSON() string {
	//Limits map[string]map[string][]Limit
	r, _ := json.Marshal(l)
	return string(r)
}
