package tdbot

import (
	"fmt"
	"log"
	"sync"

	"time"

	tdc "github.com/satyshef/go-tdlib/client"
	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/config"
	"github.com/satyshef/tdbot/profile"
	"github.com/satyshef/tdbot/user"
	"github.com/sirupsen/logrus"
)

type (
	Bot struct {
		client    *tdc.Client
		Logger    *logrus.Logger
		Profile   *profile.Profile
		Status    botStatus
		InputChan chan string
		StopWork  chan bool
		wg        sync.WaitGroup
	}
)

// New Инициализация бота
func New(prof *profile.Profile) *Bot {

	if prof == nil {
		log.Fatal("Profile nil value")
	}
	fmt.Printf("\nBot initialization %s...\n\n", prof.User.PhoneNumber)
	log := initLog(prof.Config.Log)
	tgClient := initClient(prof)
	bot := &Bot{
		client:    tgClient,
		Profile:   prof,
		Logger:    log,
		Status:    StatusStopped,
		InputChan: make(chan string, 1),
		StopWork:  make(chan bool),
	}
	/*
		time.Sleep(time.Millisecond * 100)
		bot.Logger.Infoln("Add handler ...")
		//добавляем встроенный обработчик событий
		bot.client.AddEventHandler(bot.eventCatcher)
		bot.Logger.Infoln("Handler OK")
	*/
	return bot
}

// инициализация ТГ клиента
func initClient(prof *profile.Profile) *tdc.Client {

	tdc.SetLogVerbosityLevel(1)
	tdc.SetFilePath(prof.Location() + "client.log")
	c := tdc.NewClient(
		tdc.Config{
			APIID:                  prof.Config.APP.ID,
			APIHash:                prof.Config.APP.Hash,
			SystemLanguageCode:     prof.Config.APP.SystemLanguageCode,
			DeviceModel:            prof.Config.APP.DeviceModel,
			SystemVersion:          prof.Config.APP.SystemVersion,
			ApplicationVersion:     prof.Config.APP.AppVersion,
			UseMessageDatabase:     false, //
			UseFileDatabase:        false, //
			UseChatInfoDatabase:    false, //
			UseTestDataCenter:      false,
			DatabaseDirectory:      prof.Location() + "database",
			FileDirectory:          prof.Location() + "files",
			IgnoreFileNames:        false,
			UseSecretChats:         true,
			EnableStorageOptimizer: true,
		})
	return c
}

func (bot *Bot) destroyClient() {
	if bot.client == nil {
		bot.Logger.Errorln("Client not init")
		return
	}
	//C.td_json_client_destroy(bot.client)
	bot.client.Stop()
	//time.Sleep(time.Second * 5)
	//bot.client = nil
}

func initLog(c *config.Log) *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.Level(c.Level))
	return l
}

/*
func (bot *Bot) init() {

}
*/

// Start запуск бота
func (bot *Bot) Start() *tdlib.Error {
	if bot == nil {
		return tdlib.NewError(profile.ErrorCodeNotInit, "BOT_NOT_INIT", "Bot not init")
	}
	if bot.Profile == nil {
		return tdlib.NewError(profile.ErrorCodeNotInit, "PROFILE_NOT_INIT", "Profile not init")
	}

	bot.Logger.Infoln("Starting the bot [MASTER VERSION]...")

	bot.Status = StatusInit
	bot.Profile.User.Status = user.StatusInitialization
	//fmt.Println("Log ", bot.Logger.Level)

	//добавляем встроенный обработчик событий
	bot.Logger.Infoln("Add handler ...")
	bot.client.AddEventHandler(bot.eventCatcher)
	bot.Logger.Infoln("Handler OK")

	//запуск горутины принимающей обновления телеграм
	bot.Logger.Infoln("Starting client")
	if err := bot.client.Run(); err != nil {
		fmt.Println("Start error :", err)
		bot.Stop()
		return err.(*tdlib.Error)
	}

	bot.Logger.Infoln("Client OK")
	// TODO: реализовать адекватное поведение системы если работа клиента прервалась из вне
	bot.Logger.Infoln("Init proxy ...")
	err := bot.InitProxy(true)
	if err != nil {
		bot.Logger.Errorf("PROXY ERROR %#v", err)
		if err.Message == "BOT_UNKNOWN_PROXY_TYPE" {
			return err
		}
		bot.Stop()
	}

	// TODO: тестовое отключение STOP
	err = bot.AuthBot() //2
	if err != nil {
		//bot.Stop()
		return err
	}

	// получаем инфу об аккаунте
	var me *user.User
	if me, err = bot.GetMe(true); err != nil || me.ID == 0 {
		//bot.Stop()
		return err
	}
	// устанавливаем имя из конфигурации
	//if bot.Profile.Config.APP.FirstName != "" && bot.Profile.Config.APP.FirstName != me.FirstName {
	if err := bot.SetName(bot.Profile.Config.APP.FirstName, ""); err != nil {
		//bot.Stop()
		return err
	}
	//	}

	// Устанавливаем пароль
	if err := bot.SetPassword(bot.Profile.Config.APP.AuthPass, bot.Profile.Config.APP.HintPass); err != nil {
		//bot.Stop()
		return err
	}
	// Установить фото профиля
	if err := bot.InitProfilePhoto(bot.Profile.Config.APP.Photo); err != nil {
		//bot.Stop()
		return err
	}
	// Параметры отображения номера телефона
	if err := bot.SetPhoneMode(bot.Profile.Config.APP.ShowPhoneMode); err != nil {
		//bot.Stop()
		return err
	}
	// Установить статус online
	var e error
	if bot.Profile.Config.APP.SetOnline {
		_, e = bot.client.SetOption("online", tdlib.NewOptionValueBoolean(true))
	} else {
		_, e = bot.client.SetOption("online", tdlib.NewOptionValueBoolean(false))
	}
	if e != nil {
		return err
	}

	bot.client.SetOption("ignore_background_updates", tdlib.NewOptionValueBoolean(false))
	bot.client.SetOption("ignore_sensitive_content_restrictions", tdlib.NewOptionValueBoolean(true))

	// Устанавливаем время удаления аккаунта
	if bot.Profile.Config.APP.AccountTTL != 0 {
		currentTTL, _ := bot.client.GetAccountTTL()
		if currentTTL.Days != bot.Profile.Config.APP.AccountTTL {
			bot.Logger.Debugln("Set Account TTL : ", bot.Profile.Config.APP.AccountTTL)
			ttl := tdlib.NewAccountTTL(bot.Profile.Config.APP.AccountTTL)
			_, e = bot.client.SetAccountTTL(ttl)
			if e != nil {
				return err
			}
		}
	}

	fmt.Printf("\n	Phone : %s\n	UID : %d\n	FirstName : %s\n	LastName : %s\n	UserName : %s\n	WasOnline : %s\n	Location : %s\n\n",
		bot.Profile.User.PhoneNumber,
		bot.Profile.User.ID,
		bot.Profile.User.FirstName,
		bot.Profile.User.LastName,
		bot.Profile.User.Addr,
		time.Unix(int64(bot.Profile.User.WasOnline), 0),
		bot.Profile.User.Location,
	)
	state, e := bot.client.GetAuthorizationState()
	if e != nil {
		return e.(*tdlib.Error)
	}

	for state.GetAuthorizationStateEnum() != tdlib.AuthorizationStateReadyType {
		time.Sleep(time.Second * 1)
	}

	bot.Status = StatusReady
	bot.Profile.User.Status = user.StatusReady
	ev := tdc.NewEvent(tdc.EventTypeResponse, EventNameBotReady, 0, "")
	go bot.client.PublishEvent(ev)

	<-bot.StopWork
	return nil

}

// Stop ...
func (bot *Bot) Stop() {
	//if !bot.IsRun() {
	if bot == nil || bot.client == nil || bot.Status == StatusStopped || bot.Status == StatusStopping {
		return
	}

	currentStatus := bot.Status
	bot.Logger.Infof("Stopping the bot (status %s)...\n", currentStatus)
	bot.Status = StatusStopping
	bot.Profile.User.Status = user.StatusStopped

	// Если не дождались WG тогда принудительно останавливаем(наблюдались зависания при TIMEOUT запроса)
	stop := false
	timeout := 20
	go func() {
		for i := 0; i <= timeout; i++ {
			if stop {
				return
			}
			time.Sleep(time.Second * 1)
		}
		if bot != nil {
			bot.wg.Done()
		}
	}()

	bot.wg.Wait()
	stop = true
	bot.destroyClient()
	bot.Status = StatusStopped
	if currentStatus == StatusReady {
		bot.StopWork <- true
	}
	//bot.Profile.Close()
	bot.Logger.Infoln("Stopping finish")

}

// DONT USED!
func (bot *Bot) Restart1() {

	if bot.Profile == nil {
		bot.Logger.Debugln("Bot not restarted : profile not init")
		return
	}

	bot.Logger.Infoln("Restarting the bot ...")
	//	bot.Status = StatusRestart

	bot.Stop()
	//time.Sleep(time.Second * 2)
	bot.Start()

	bot.Logger.Infoln("Restarting finish")
}

func (bot *Bot) IsRun() bool {
	//if bot == nil || bot.Status == StatusStopped || bot.Status == StatusStopping || bot.client == nil {
	if bot == nil || bot.Status != StatusReady || bot.client == nil {
		return false
	}
	return true
}

func (bot *Bot) IsClientInit() bool {
	if bot.client == nil {
		return false
	}
	return true
}

// Init proxy server
func (bot *Bot) InitProxy(check bool) *tdlib.Error {
	/*
		if !bot.IsRun() {
			return tdlib.NewError(ErrorCodeSystem, "BOT_DONT_READY", "")
		}
	*/
	/*
		if !bot.isReady() {
			return tdlib.NewError(ErrorSystem, "BOT_NOT_READY", "")
		}
	*/

	if bot.Profile == (&profile.Profile{}) || bot.Profile.Config == nil {
		return tdlib.NewError(profile.ErrorCodeNotInit, "PROFILE_NOT_INIT", "Profile not init")
	}

	if bot.Profile.Config.Proxy == nil {
		bot.Logger.Infoln("Proxy not set")
		return nil
	}

	if bot.Profile.Config.Proxy != nil && !bot.Profile.Config.Proxy.Enable {
		bot.Logger.Infoln("Disabling Proxies ...")
		_, err := bot.client.DisableProxy()
		/*
			//Если установлен параметр disable удаляем все прокси(ОБДУМАТЬ)

			err := bot.RemoveAllProxy()
		*/
		if err != nil {
			bot.Logger.Errorf("Proxy not disable : %#v\n", err)
			return err.(*tdlib.Error)
		}
		return nil
	} else {

		var prox *tdlib.Proxy
		var proxType tdlib.ProxyType

		switch bot.Profile.Config.Proxy.Type {
		case config.ProxyTypeSocks5:
			// Socks5
			proxType = tdlib.NewProxyTypeSocks5(bot.Profile.Config.Proxy.User, bot.Profile.Config.Proxy.Pass)
		case config.ProxyTypeHttp:
			// HTTP - HTTPS proxy
			proxType = tdlib.NewProxyTypeHttp(bot.Profile.Config.Proxy.User, bot.Profile.Config.Proxy.Pass, false)
		case config.ProxyTypeMtproto:
			// MtProto Proxy
			proxType = tdlib.NewProxyTypeMtproto(bot.Profile.Config.Proxy.Pass)
		default:
			return tdlib.NewError(ErrorCodeSystem, "BOT_UNKNOWN_PROXY_TYPE", fmt.Sprintf("Unknown proxy type : %s", bot.Profile.Config.Proxy.Type))
		}

		/*
			//Если прокси уже есть то его не добавляем
			var err *tdlib.Error
			prox, err = bot.FindProxy(bot.Profile.Config.Proxy.Host, bot.Profile.Config.Proxy.Port, proxType)
			if err != nil {
				if err.Code == 404 {
					var e error
					bot.Logger.Infoln("Add New Proxy")
					prox, e = bot.client.AddProxy(bot.Profile.Config.Proxy.Host, bot.Profile.Config.Proxy.Port, true, proxType)
					if e != nil {
						return e.(*tdlib.Error)
					}
				} else {
					return err
				}
			}

		*/

		/*
			proxyList, _ := bot.client.GetProxies()
			fmt.Printf("Proxy List %#v\n", proxyList)
		*/

		prox, e := bot.client.AddProxy(bot.Profile.Config.Proxy.Host, bot.Profile.Config.Proxy.Port, true, proxType)
		if e != nil {
			return e.(*tdlib.Error)
		}

		_, e = bot.client.EnableProxy(prox.ID)
		if e != nil {
			return e.(*tdlib.Error)
		}

		bot.Logger.Infoln("Set Proxy ", prox.Server)
		//bot.Logger.Infoln("Proxy ID", prox.ID)

		if check {
			ping, e := bot.client.PingProxy(prox.ID)
			if e != nil {
				return e.(*tdlib.Error)
			}

			bot.Logger.Infoln("Ping ", ping.Seconds)
		}

	}

	return nil
}

// FindProxy проверить есть ли прокси в списке подключенных
func (bot *Bot) FindProxy(host string, port int32, proxyType tdlib.ProxyType) (*tdlib.Proxy, *tdlib.Error) {

	proxyList, err := bot.client.GetProxies()
	//fmt.Printf("Proxy List %#v\n", proxyList)
	if err != nil {
		return &tdlib.Proxy{}, err.(*tdlib.Error)
	}

	for _, pr := range proxyList.Proxies {
		if pr.Server == host && pr.Port == port && pr.Type.GetProxyTypeEnum() == proxyType.GetProxyTypeEnum() {
			return &pr, nil
		}
	}

	return &tdlib.Proxy{}, tdlib.NewError(tdc.ErrorCodeNotFound, "Proxy not found", "")
}

func (bot *Bot) RemoveAllProxy() *tdlib.Error {
	proxyList, err := bot.client.GetProxies()
	//fmt.Printf("Proxy List %#v\n", proxyList)
	if err != nil {
		return err.(*tdlib.Error)
	}

	for _, pr := range proxyList.Proxies {

		_, err := bot.client.RemoveProxy(pr.ID)
		if err != nil {
			return err.(*tdlib.Error)
		}
	}

	return nil
}

// старт процесса авторизации
func (bot *Bot) AuthBot() *tdlib.Error {

	for {
		if bot.Status == StatusStopped || bot.Status == StatusStopping {
			return tdlib.NewError(tdc.ErrorCodeAborted, "CLIENT_ABORTED", "Client authorization interrupted")
		}

		currentState, err := bot.client.Authorize()
		if err != nil {
			//bot.Logger.Error("Authorization error : ", err)
			return err.(*tdlib.Error)
		}

		if currentState == nil {
			bot.Logger.Infoln("Sleep")
			time.Sleep(time.Second * 1)
			continue
		}

		//bot.Logger.Errorln(bot.Profile.User.Status)
		/*
			if bot.Profile.User.Status == user.StatusBanned {
				return tdlib.NewError(ErrorSystem, "Bot banned", "")
			}
		*/
		switch currentState.GetAuthorizationStateEnum() {
		/*
			case tdlib.AuthorizationStateWaitTdlibParametersType:

					bot.Restart()
					return
		*/

		case tdlib.AuthorizationStateWaitEncryptionKeyType:
			//bot.Profile.User.Status = user.StatusInitialization
			bot.Logger.Infoln("Wait network....")
			continue

		case tdlib.AuthorizationStateWaitPhoneNumberType:
			//Если в профиле указан номер телефона то отправляем его
			if bot.Profile.User.PhoneNumber == "" {
				//	bot.Profile.User.Status = user.StatusWaitPhone
				bot.Logger.Info("Wait phone number")
				bot.Profile.User.PhoneNumber = <-bot.InputChan
			}
			//bot.Profile.User.Status = user.StatusSendPhone
			bot.Logger.Infoln("Send phone ...")

			if _, err := bot.client.SendPhoneNumber(bot.Profile.User.PhoneNumber); err != nil {
				//bot.Logger.Errorf("SEND PHONE ERROR : %#v\n", err.(*tdlib.Error))
				//fmt.Printf("SEND PHONE ERROR (user status %s): %s\n", bot.Profile.User.Status, err.Error())
				return err.(*tdlib.Error)
				//time.Sleep(time.Second * 3)
				//continue
			}

		case tdlib.AuthorizationStateWaitCodeType:
			//Если код был отправлен в телеграм то дублируем его в смс
			s := currentState.(*tdlib.AuthorizationStateWaitCode)
			if s.CodeInfo.Type != nil && s.CodeInfo.Type.GetAuthenticationCodeTypeEnum() == tdlib.AuthenticationCodeTypeTelegramMessageType {
				bot.Logger.Infof("Code send to Telegram\n")
			} else {
				bot.Logger.Infof("Code send to SMS \n")
			}
			//bot.Profile.User.Status = user.StatusWaitCode
			bot.Logger.Info("Wait code ...")
			_, err := bot.client.SendAuthCode(<-bot.InputChan)
			if err != nil {
				bot.Logger.Errorf("Wait Code : %s\n", err)
				continue
			}

		case tdlib.AuthorizationStateWaitPasswordType:
			var pass string
			//если установлен пароль в настройка то пробуем его применить
			if bot.Profile.Config.APP.AuthPass != "" {
				pass = bot.Profile.Config.APP.AuthPass
			} else {
				//bot.Profile.User.Status = user.StatusWaitPassword
				s := currentState.(*tdlib.AuthorizationStateWaitPassword)
				bot.Logger.Infof("Wait password. Hint: %s", s.PasswordHint)
				pass = <-bot.InputChan
			}
			_, err := bot.client.SendAuthPassword(pass)
			if err != nil {
				// TODO: Тест с возвращением ошибки, в рабочем варианте продолжал работать
				return err.(*tdlib.Error)
				/*
					bot.Logger.Errorf("Send password error %s\n", err.Error())
					continue
				*/
				//TEST!!!!
				/*
					if err.(*tdlib.Error).Code == tdlib.ErrorCodePassInvalid && bot.Profile.Config.APP.AuthPass != "" {
						bot.Profile.Config.APP.AuthPass = ""
					}
				*/

			}

		case tdlib.AuthorizationStateWaitRegistrationType:
			bot.Logger.Infof("Register new user ....")
			//bot.Profile.User.Status = user.StatusRegistration
			//Устанавливаем имя пользователя
			var firstName string
			if bot.Profile.Config.APP.FirstName != "" {
				firstName = bot.Profile.Config.APP.FirstName
			} else {
				firstName = bot.Profile.User.PhoneNumber
			}
			ok, err := bot.client.RegisterUser(firstName, "")
			if err != nil {
				bot.Logger.Errorf("Register user %s\n", err)
				//os.Exit(1)
				return err.(*tdlib.Error)
			}
			bot.Logger.Infof("Register : %s\n", ok.Type)

		case tdlib.AuthorizationStateReadyType:
			bot.Logger.Info("Authorization Ready! Let's rock")
			//bot.Profile.User.Status = user.StatusReady
			goto Ready

		case tdlib.AuthorizationStateLoggingOutType:
			//bot.Profile.User.Status = user.StatusLogout
			return tdlib.NewError(tdc.ErrorCodeLogout, "CLIENT_LOGOUT", "Client logout")

		case tdlib.AuthorizationStateClosedType:
			return tdlib.NewError(tdc.ErrorCodeSystem, "CLIENT_SYSTEM_ERROR", "Authorization State Closed")

		default:
			fmt.Printf("Switch Default : %s\n", currentState.GetAuthorizationStateEnum())
			time.Sleep(time.Second * 1)
			continue
		}

	}

Ready:
	return nil
}

func (bot *Bot) GetRawUpdatesChannel(capacity int) (chan *tdlib.UpdateMsg, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	return bot.client.GetRawUpdatesChannel(capacity), nil
}

func (bot *Bot) AddEventHandler(event tdc.EventHandler) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	bot.client.AddEventHandler(event)
	return nil
}

func (bot *Bot) GetMarkdownText(text *tdlib.FormattedText) (*tdlib.FormattedText, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}

	result, err := bot.client.GetMarkdownText(text)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	return result, nil
}
