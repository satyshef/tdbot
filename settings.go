package tdbot

import (
	"strings"
	"time"

	tdc "github.com/satyshef/go-tdlib/client"
	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/events/event"
	"github.com/satyshef/tdbot/profile"
	"github.com/satyshef/tdbot/user"
)

// SendPhone ....
func (bot *Bot) SendPhone(phoneNumber string) {
	bot.InputChan <- phoneNumber
	/*
		//Что делать если номер отличается от записанного в профиле???
		bot.Logger.Infof("Send phone number %s....", phoneNumber)
		_, err := bot.client.SendPhoneNumber(phoneNumber)
		//bot.mutex.Unlock()
		if err != nil {
			e := err.(*tdlib.Error)
			bot.Logger.Errorf("Error sending phone number : %s\n", e.Message)
			return e
		}

		return nil
	*/
}

// SendCode ....
func (bot *Bot) SendCode(code string) {
	bot.InputChan <- code
	/*
		_, err := bot.client.SendAuthCode(code)

		if err == nil {
			bot.mutex.Unlock()
			return nil
		}

		//если таймаут не делаем unlock
		e := err.(*tdlib.Error)
		if e.Code != tdlib.ErrorCodeTimeout {
			bot.mutex.Unlock()
		}
		return e
	*/
}

func (bot *Bot) ResendCode() (e *tdlib.Error) {
	_, err := bot.client.ResendAuthenticationCode()
	if err != nil {
		e = err.(*tdlib.Error)
	}

	//bot.mutex.Unlock()
	return e

}

// SendPass ....
func (bot *Bot) SendPass(pass string) {
	bot.InputChan <- pass
	/*
		_, err := bot.client.SendAuthPassword(pass)
		if err != nil {
			e := err.(*tdlib.Error)
			return e
		}

		//bot.mutex.Unlock()
		return nil
	*/
}

func (bot *Bot) SetPassword(pass, hint string) *tdlib.Error {
	passState, e := bot.client.GetPasswordState()
	if e != nil {
		return e.(*tdlib.Error)
	}

	if !passState.HasPassword && pass != "" {
		var err error
		passState, err = bot.client.SetPassword("", pass, hint, false, "")
		if err != nil {
			//bot.Logger.Errorf("Set pass %#v\n", err)
			return err.(*tdlib.Error)
		} else {
			bot.Logger.Infof("Set new pass : OK \n")
		}
	}

	return nil
}

// SetPhoneMode установить режим отображения собственного номера телефона
func (bot *Bot) SetPhoneMode(mode int) *tdlib.Error {

	//скрываем номер телефона
	var pr tdlib.UserPrivacySettingRule

	switch mode {
	case 1:
		// Показывать номер телефона всем
		pr = tdlib.NewUserPrivacySettingRuleAllowAll()
	case 2:
		// Скрывать номер для всех
		pr = tdlib.NewUserPrivacySettingRuleRestrictAll()
	default:
		return nil
	}

	ps := tdlib.NewUserPrivacySettingShowPhoneNumber()
	_, err := bot.client.SetUserPrivacySettingRules(ps, tdlib.NewUserPrivacySettingRules([]tdlib.UserPrivacySettingRule{pr}))

	if err != nil {
		return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
	} else {
		bot.Logger.Infof("Установлены настройки отображения номера телефона : %s\n", pr.GetUserPrivacySettingRuleEnum())
	}

	return nil
}

// GetMe Информация о текущем пользователе
func (bot *Bot) GetMe(reload bool) (*user.User, *tdlib.Error) {
	if !reload {
		return bot.Profile.User, nil
	}
	var usr *user.User
	me, err := bot.client.GetMe()
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	/*
		fmt.Printf("%#v\n\n", l)
		fmt.Printf("%#v\n\n", me)
	*/
	usr = ConvertUser(me)

	//Страна бота
	country, err := bot.client.GetCountryCode()
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	usr.Location = country.Text
	usr.Status = bot.Profile.User.Status
	usr.Role = bot.Profile.Config.Condition.Role
	usr.Group = bot.Profile.Config.Condition.Group
	bot.Profile.User = usr
	return usr, nil
}

// SetName установить имя пользователя
func (bot *Bot) SetName(firstname, lastname string) *tdlib.Error {
	firstname = strings.Title(firstname)
	lastname = strings.Title(lastname)
	if bot.Profile.User.FirstName == firstname && bot.Profile.User.LastName == lastname {
		//bot.Logger.Infof("The name match")
		return nil
	}

	if _, err := bot.client.SetName(firstname, lastname); err != nil {
		return err.(*tdlib.Error)
	}

	bot.Profile.User.FirstName = firstname
	bot.Profile.User.LastName = lastname

	bot.Logger.Infof("Set name success : %s %s", firstname, lastname)

	return nil
}

// ======================== Profile Methods =====================================

func (bot *Bot) InitProfilePhoto(photo string) *tdlib.Error {
	if photo == "remove" {
		err := bot.ProfileRemovePhoto()
		if err != nil {
			//bot.Logger.Errorf("Не удалось удалить фото профиля : %s\n", err)
			return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
		} else {
			bot.Logger.Infof("Delete photo success\n")
		}
	} else if photo != "" {
		err := bot.ProfileSetPhoto(bot.Profile.Config.APP.Photo)
		if err != nil {
			return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
		} else {
			bot.Logger.Infof("Set profile photo %s success\n", bot.Profile.Config.APP.Photo)
		}
	}

	return nil
}

// SetProfilePhoto установить фото профиля
func (bot *Bot) ProfileSetPhoto(path string) *tdlib.Error {
	inFile := tdlib.NewInputFileLocal(path)
	chatPhoto := tdlib.NewInputChatPhotoStatic(inFile)
	_, err := bot.client.SetProfilePhoto(chatPhoto)
	if err != nil {
		return err.(*tdlib.Error)
	}

	return nil

}

// RemoveProfilePhoto удалить фото профиля
func (bot *Bot) ProfileRemovePhoto() *tdlib.Error {
	p, err := bot.client.GetUserProfilePhotos(bot.Profile.User.ID, 0, 100)
	if err != nil {
		bot.Logger.Errorf("REMOVE PROFILE %#v", err)
		return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
		//return err.(*tdlib.Error)
	} else {
		for _, photo := range p.Photos {
			_, err = bot.client.DeleteProfilePhoto(&photo.ID)
			if err != nil {
				return err.(*tdlib.Error)
			}
		}
	}

	return nil
}

func (b *Bot) ProfileToSpam() *tdlib.Error {
	if !b.IsRun() {
		return tdlib.NewError(tdc.ErrorCodeSystem, "BOT_IS_DOWN", "")
	}
	if b.Profile.Config.APP.DirFoul == "" {
		return tdlib.NewError(tdc.ErrorCodeSystem, "Foul dir not set", "")
	}
	b.Stop()
	var path string
	if b.Profile.Config.APP.DirFoul[0:1] == "/" {
		path = b.Profile.Config.APP.DirFoul
	} else {
		path = b.Profile.BaseDir() + b.Profile.Config.APP.DirFoul + "/" + time.Now().Format("2006-01-02")
	}
	err := b.Profile.Move(path)
	if err != nil {
		return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
	}
	b.Profile.Close(1000)
	b.Profile = nil
	return nil
}

func (b *Bot) ProfileToLogout() error {

	/*
		if !b.IsRun() {
			return tdlib.NewError(tdc.ErrorCodeSystem, "BOT_IS_DOWN", "")
		}
	*/

	if b.Profile.Config.APP.DirLogout == "" {
		return tdlib.NewError(tdc.ErrorCodeSystem, "Logout dir not set", "")
	}
	b.Stop()
	var path string
	if b.Profile.Config.APP.DirLogout[0:1] == "/" {
		path = b.Profile.Config.APP.DirLogout
	} else {
		path = b.Profile.BaseDir() + b.Profile.Config.APP.DirLogout + "/" + time.Now().Format("2006-01-02")
	}
	err := b.Profile.Move(path)
	if err != nil {
		return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
	}
	b.Profile.Close(1000)
	b.Profile = nil
	return nil
}

func (b *Bot) ProfileToBan() error {
	if !b.IsRun() {
		return tdlib.NewError(tdc.ErrorCodeSystem, "BOT_IS_DOWN", "")
	}
	if b.Profile.Config.APP.DirBanned == "" {
		return tdlib.NewError(tdc.ErrorCodeSystem, "Ban dir not set", "")
	}
	b.Stop()
	var path string
	if b.Profile.Config.APP.DirBanned[0:1] == "/" {
		path = b.Profile.Config.APP.DirBanned
	} else {
		path = b.Profile.BaseDir() + b.Profile.Config.APP.DirBanned + "/" + time.Now().Format("2006-01-02")
	}
	err := b.Profile.Move(path)
	if err != nil {
		return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
	}
	b.Profile.Close(1000)
	b.Profile = nil
	return nil
}

func (b *Bot) ProfileToDouble() error {
	if !b.IsRun() {
		return tdlib.NewError(tdc.ErrorCodeSystem, "BOT_IS_DOWN", "")
	}
	if b.Profile.Config.APP.DirDouble == "" {
		return tdlib.NewError(tdc.ErrorCodeSystem, "Double dir not set", "")
	}
	b.Stop()
	var path string
	if b.Profile.Config.APP.DirDouble[0:1] == "/" {
		path = b.Profile.Config.APP.DirDouble
	} else {
		path = b.Profile.BaseDir() + b.Profile.Config.APP.DirDouble + "/" + time.Now().Format("2006-01-02")
	}
	err := b.Profile.Move(path)
	if err != nil {
		return tdlib.NewError(tdc.ErrorCodeSystem, err.Error(), "")
	}
	b.Profile.Close(1000)
	b.Profile = nil
	return nil
}

// Проверяем лимиты события
func (bot *Bot) CheckEventLimits(evnt *event.Event) *tdlib.Error {
	if !bot.Profile.Config.APP.CheckLimits {
		return nil
	}
	needStop := false
	//bot.Logger.Errorln("Check LImit ", eventType, eventName)
	exLimits := bot.Profile.CheckLimit(evnt.Type, evnt.Name)
	if len(exLimits) == 0 {
		return nil
	}
	bot.Logger.Debugf("======== LIMIT %s :: %s ===========\n", evnt.Type, evnt.Name)
	for _, limit := range exLimits {
		//если до оканачания ограничений много времени тогда останавливаем бота
		//if limit.Interval > bot.Profile.Config.APP.DontRebootInterval && bot.Profile.Config.APP.Mode == 2 {
		if limit.Interval > bot.Profile.Config.APP.DontRebootInterval {
			needStop = true
		}
		//l := &config.Limits{evnt.Type: {evnt.Name: exLimits}}
		bot.Logger.Debugln(limit.Limit, " : ", limit.Interval)
	}

	if needStop {
		bot.Stop()
	}
	return tdlib.NewError(profile.ErrorCodeLimitExceeded, "BOT_LIMIT_EXCEEDED", "")
}
