package tdbot

import (
	"strings"

	"github.com/polihoster/tdbot/config"
	"github.com/polihoster/tdbot/events/event"
	"github.com/polihoster/tdbot/profile"
	"github.com/polihoster/tdlib"
)

// обработчик событий Телеграм клиента
func (bot *Bot) eventCatcher(eventType string, update interface{}) *tdlib.Error {
	//bot.Logger.Infof("NEW EVENT %s: %#v\n", eventType, update)

	if update == nil {
		return tdlib.NewError(tdlib.ErrorCodeSystem, "CLIENT_EMPTY_UPDATE", "Received an empty update to client")
	}

	//Эксперемент
	if bot.Profile == nil {
		go bot.Restart()
		return tdlib.NewError(profile.ErrorCodeDirNotExists, "PROFILE_NOT_INIT", "Profile not init")
	}

	if bot.Client == nil {
		return tdlib.NewError(tdlib.ErrorCodeNotInit, "CLIENT_NOT_INIT", "Client not init")
	}

	// Проверяем существование директории профиля. Решить что делать при отсутcтвии директории
	if !bot.Profile.CheckDir() {
		//bot.Profile.User.Status = user.StatusNotExist
		//go bot.Stop()
		return tdlib.NewError(profile.ErrorCodeDirNotExists, "PROFILE_DIR_NOT_EXISTS", "Profile Dir Not Exists")
	}

	switch eventType {
	case tdlib.EventTypeRequest:
		return bot.requestHandler(update.(tdlib.UpdateMsg))
	case tdlib.EventTypeResponse:
		return bot.responseHandler(update.(tdlib.UpdateMsg))
	case tdlib.EventTypeError:
		return bot.errorHandler(update.(tdlib.Error))
	}

	return nil
}

// оброботчик запросов к серверу Телеграм
func (bot *Bot) requestHandler(request tdlib.UpdateMsg) *tdlib.Error {
	//bot.Logger.Errorf("New Request %#v\n\n", request.Data)

	requestName := request.Data["@type"].(string)
	if err := checkEventLimits(bot, profile.EventTypeRequest, requestName); err != nil {
		return err
	}

	ev := event.New(profile.EventTypeRequest, requestName, 0, string(request.Raw))
	if err := bot.Profile.Event.Write(ev); err != nil && !strings.Contains(err.Error(), "Event not observed") {
		bot.Logger.Errorln(err)
		bot.Stop()
		return tdlib.NewError(profile.ErrorCodeLimitExceeded, "PROFILE_EVENT_DONT_WRITE", err.Error())
	}

	return nil
}

//Обработчик ответов на запросы Telegram client (не используется)
func (bot *Bot) responseHandler(response tdlib.UpdateMsg) *tdlib.Error {
	//bot.Logger.Infof("Response %#v\n\n", response.Data)

	respName := response.Data["@type"].(string)
	ev := event.New(profile.EventTypeResponse, respName, 0, string(response.Raw))
	if err := bot.Profile.Event.Write(ev); err != nil && strings.Contains(err.Error(), "Event not observed") {
		bot.Logger.Errorln(err)
		bot.Stop()
		return tdlib.NewError(profile.ErrorCodeLimitExceeded, "PROFILE_EVENT_DONT_WRITE", err.Error())
	}

	// Проверяем лимиты ответа
	if err := checkEventLimits(bot, profile.EventTypeResponse, respName); err != nil {
		return err
	}
	return nil
}

//Обработчик ошибок Telegram client. Пересмотреть логику функции
func (bot *Bot) errorHandler(e tdlib.Error) *tdlib.Error {
	//bot.Logger.Errorf("Error Catch %#v\n\n", e)

	ev := event.New(profile.EventTypeError, e.Type, 0, e.Message)
	if err := bot.Profile.Event.Write(ev); err != nil && strings.Contains(err.Error(), "Event not observed") {
		bot.Logger.Errorln(err)
		bot.Stop()
		return tdlib.NewError(profile.ErrorCodeLimitExceeded, "PROFILE_EVENT_DONT_WRITE", err.Error())
	}

	// Проверяем лимиты ошибки
	if err := checkEventLimits(bot, profile.EventTypeError, e.Type); err != nil {
		return err
	}

	return nil
}

// Проверяем лимиты события
func checkEventLimits(bot *Bot, eventType, eventName string) *tdlib.Error {
	//bot.Logger.Errorln("Check LImit ", eventType, eventName)
	exLimits := bot.Profile.CheckLimit(eventType, eventName)

	for _, limit := range exLimits {
		//если до оканачания ограничений много времени тогда останавливаем бота
		if limit.Interval > bot.Profile.Config.APP.DontRebootInterval && bot.Profile.Config.APP.Mode == 2 {
			bot.Stop()
		}

		l := &config.Limits{eventType: {eventName: exLimits}}
		return tdlib.NewError(profile.ErrorCodeLimitExceeded, "BOT_LIMIT_EXCEEDED", l.JSON())
	}

	return nil
}
