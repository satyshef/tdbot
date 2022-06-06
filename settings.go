package tdbot

import (
	"github.com/satyshef/tdbot/config"
	"github.com/satyshef/tdbot/events/event"
	"github.com/satyshef/tdbot/profile"
	"github.com/satyshef/tdbot/user"
	"github.com/satyshef/tdlib"
)

// SendPhone ....
func (bot *Bot) SendPhone(phoneNumber string) {
	bot.InputChan <- phoneNumber
	/*
		//Что делать если номер отличается от записанного в профиле???
		bot.Logger.Infof("Send phone number %s....", phoneNumber)
		_, err := bot.Client.SendPhoneNumber(phoneNumber)
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
		_, err := bot.Client.SendAuthCode(code)

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
	_, err := bot.Client.ResendAuthenticationCode()
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
		_, err := bot.Client.SendAuthPassword(pass)
		if err != nil {
			e := err.(*tdlib.Error)
			return e
		}

		//bot.mutex.Unlock()
		return nil
	*/
}

func (bot *Bot) SetPassword(pass, hint string) *tdlib.Error {
	passState, e := bot.Client.GetPasswordState()
	if e != nil {
		return e.(*tdlib.Error)
	}

	if !passState.HasPassword && pass != "" {
		var err error
		passState, err = bot.Client.SetPassword("", pass, hint, false, "")
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
	_, err := bot.Client.SetUserPrivacySettingRules(ps, tdlib.NewUserPrivacySettingRules([]tdlib.UserPrivacySettingRule{pr}))

	if err != nil {
		return tdlib.NewError(tdlib.ErrorCodeSystem, err.Error(), "")
	} else {
		bot.Logger.Infof("Установлены настройки отображения номера телефона : %s\n", pr.GetUserPrivacySettingRuleEnum())
	}

	return nil
}

//GetMe Информация о текущем пользователе
func (bot *Bot) GetMe() (*user.User, *tdlib.Error) {
	var usr *user.User
	me, err := bot.Client.GetMe()
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	usr = ConvertUser(me)

	//Страна бота
	country, err := bot.Client.GetCountryCode()
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	usr.Location = country.Text
	usr.Status = bot.Profile.User.Status
	bot.Profile.User = usr
	return usr, nil
}

// SetName установить имя пользователя
func (bot *Bot) SetName(firstname, lastname string) *tdlib.Error {
	if bot.Profile.User.FirstName == firstname && bot.Profile.User.LastName == lastname {
		bot.Logger.Infof("The name match")
		return nil
	}

	if _, err := bot.Client.SetName(firstname, lastname); err != nil {
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
			return tdlib.NewError(tdlib.ErrorCodeSystem, err.Error(), "")
		} else {
			bot.Logger.Infof("Delete photo success\n")
		}
	} else if photo != "" {
		err := bot.ProfileSetPhoto(bot.Profile.Config.APP.Photo)
		if err != nil {
			return tdlib.NewError(tdlib.ErrorCodeSystem, err.Error(), "")
		} else {
			bot.Logger.Infof("Set profile photo %s success\n", bot.Profile.Config.APP.Photo)
		}
	}

	return nil
}

//SetProfilePhoto установить фото профиля
func (bot *Bot) ProfileSetPhoto(path string) *tdlib.Error {
	inFile := tdlib.NewInputFileLocal(path)
	chatPhoto := tdlib.NewInputChatPhotoStatic(inFile)
	_, err := bot.Client.SetProfilePhoto(chatPhoto)
	if err != nil {
		return err.(*tdlib.Error)
	}

	return nil

}

//RemoveProfilePhoto удалить фото профиля
func (bot *Bot) ProfileRemovePhoto() *tdlib.Error {
	p, err := bot.Client.GetUserProfilePhotos(bot.Profile.User.ID, 0, 100)
	if err != nil {
		bot.Logger.Errorf("REMOVE PROFILE %#v", err)
		return tdlib.NewError(tdlib.ErrorCodeSystem, err.Error(), "")
		//return err.(*tdlib.Error)
	} else {
		for _, photo := range p.Photos {
			_, err = bot.Client.DeleteProfilePhoto(photo.ID)
			if err != nil {
				return err.(*tdlib.Error)
			}
		}
	}

	return nil
}

func (b *Bot) ProfileToSpam() error {
	b.Stop()
	return b.Profile.Move(b.Profile.BaseDir() + "spam")
}

func (b *Bot) ProfileToLogout() error {
	b.Stop()
	return b.Profile.Move(b.Profile.BaseDir() + "logout")
}

func (b *Bot) ProfileToBan() error {
	b.Stop()
	return b.Profile.Move(b.Profile.BaseDir() + "banned")
}

// Проверяем лимиты события
func (bot *Bot) CheckEventLimits(evnt *event.Event) *tdlib.Error {
	//bot.Logger.Errorln("Check LImit ", eventType, eventName)
	exLimits := bot.Profile.CheckLimit(evnt.Type, evnt.Name)
	for _, limit := range exLimits {
		//если до оканачания ограничений много времени тогда останавливаем бота
		if limit.Interval > bot.Profile.Config.APP.DontRebootInterval && bot.Profile.Config.APP.Mode == 2 {
			bot.Stop()
		}
		l := &config.Limits{evnt.Type: {evnt.Name: exLimits}}
		return tdlib.NewError(profile.ErrorCodeLimitExceeded, "BOT_LIMIT_EXCEEDED", l.JSON())
	}
	return nil
}
