package tdbot

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/polihoster/tdlib"
)

/*
func (bot *Bot) SearchChats(chatName string) (*tdlib.Chats, *tdlib.Error) {

	//chats, err := bot.Client.SearchChats(chatName, 1000)
	chats, err := bot.Client.SearchPublicChats(chatName)

	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	for _, cid := range chats.ChatIDs {

		chat, _ := bot.Client.GetChat(cid)
		c := chat.Type.(*tdlib.ChatTypeSupergroup)
		cc, _ := bot.Client.GetSupergroup(c.SupergroupID)
		bot.Logger.Infof("%#v\n\n", cc)
	}

	return chats, nil
}

*/

func (bot *Bot) InviteUser(uid int32, cid string) *tdlib.Error {

	destChat, e := bot.GetChat(cid, true)
	if e != nil {
		return e
	}

	bot.Logger.Infoln("Add user ID", uid)
	_, err := bot.Client.AddChatMember(destChat.ID, uid, 100)
	if err != nil {
		bot.Logger.Error("%#v\n", err)
		return err.(*tdlib.Error)
	} else {
		bot.Logger.Infoln("Add OK")
	}

	return nil
}

//Получить полную информацию о пользователе
func (bot *Bot) GetUser(uid int32, cid string) (*tdlib.UserFullInfo, *tdlib.Error) {
	if cid != "" {
		//_, e := bot.GetChat(cid, false)
		_, e := bot.GetChatMembers(cid, 0, 0)
		if e != nil {
			return nil, e
		}
		//	fmt.Printf("user %#v\n", m)

	}

	user, err := bot.Client.GetUserFullInfo(uid)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	fmt.Printf("user %#v\n", user)

	return user, nil
}

// Отправить сообщения участникам группы
func (bot *Bot) SendMessageChatMembers(chatID string, message string, offset int32, limit int32) ([]tdlib.ChatMember, *tdlib.Error) {

	result := []tdlib.ChatMember{}
	members, err := bot.GetChatMembers(chatID, offset, limit)

	if len(members) > 0 {
		for _, m := range members {
			_, err := bot.SendMessageByUID(m.UserID, message, 0)
			time.Sleep(time.Second * 2)
			if err != nil {
				if err.Code == tdlib.ErrorCodeNoAccess {
					bot.Logger.Errorf("%d - %s", m.UserID, err.Message)
					continue
				}
				return result, err
			}
			result = append(result, m)
			bot.Logger.Printf("[ %s ] Send to user ID %d - OK\n", time.Now().Format(time.RFC3339), m.UserID)

		}
	}
	return result, err

}

// Загрузить полную информацию об участниках группы
func (bot *Bot) GetChatUsers(chatID string, offset int32, limit int32) ([]tdlib.User, *tdlib.Error) {

	result := []tdlib.User{}

	members, err := bot.GetChatMembers(chatID, offset, limit)
	if len(members) > 0 {
		for _, m := range members {
			user, err := bot.Client.GetUser(m.UserID)
			if err != nil {
				bot.Logger.Error("Get User", err)
			} else {
				result = append(result, *user)
			}
		}
	}
	return result, err

}

// Скопировать участников одной группы в другую
func (bot *Bot) CopyChatUsers(source, destination string, offset int32, limit int32) ([]tdlib.ChatMember, *tdlib.Error) {

	result := []tdlib.ChatMember{}

	// Загружаем список участников группы-донора
	members, err := bot.GetChatMembers(source, offset, limit)
	if len(members) > 0 {
		destChat, e := bot.GetChat(destination, true)
		if e != nil {
			return nil, e
		}

		// Добавляем участников в целевую группу
		for _, m := range members {
			// Исключаем себя
			if bot.Profile.User.ID == m.UserID {
				continue
			}

			user, err := bot.Client.GetUser(m.UserID)
			if err != nil {
				bot.Logger.Errorln("Get User Error ", err)
				continue
			}

			//Фильтр только живые люди
			// Пропускаем скамеров, ботов, администраторов
			if user.IsScam || user.Type.GetUserTypeEnum() == tdlib.NewUserTypeDeleted().GetUserTypeEnum() || user.Type.GetUserTypeEnum() == tdlib.NewUserTypeBot(false, false, false, "", false).GetUserTypeEnum() {
				continue
			}

			bot.Logger.Infoln("Add user ID", m.UserID)
			_, err = bot.Client.AddChatMember(destChat.ID, m.UserID, 100)
			if err != nil {
				//Если заблокировали за флуд возвращаем результат
				if err.(*tdlib.Error).Code == tdlib.ErrorCodeFloodLock {
					return result, err.(*tdlib.Error)
				}
				bot.Logger.Error("AddChatMember", err)
			} else {
				result = append(result, m)
			}
		}
	}
	return result, err

}

// Загрузить список участников группы. Группа может быть как обычная(до 200) так и супергруппа(до 200к) участников. Возвращает ID участников и их статус
// @chatID - имя или сслыка на чат
func (bot *Bot) GetChatMembers(chatID string, offset int32, limit int32) ([]tdlib.ChatMember, *tdlib.Error) {
	var chat *tdlib.Chat
	var err error
	var result []tdlib.ChatMember

	if limit < 0 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Limit must be positive")
	}

	if offset < 0 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Offset must be positive")
	}

	chat, e := bot.GetChat(chatID, false)
	if e != nil {
		return nil, e
	}

	bot.Logger.Infoln("Parsing chat : ", chatID)

	// Загружаем участников чата
	if chat.Type.GetChatTypeEnum() == tdlib.ChatTypeSupergroupType {
		c := chat.Type.(*tdlib.ChatTypeSupergroup)
		if limit == 0 {
			s, err := bot.Client.GetSupergroup(c.SupergroupID)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
			limit = s.MemberCount - offset
		}

		if limit > 10000 {
			return result, tdlib.NewError(ErrorCodeSystem, "BOT_LIMIT_EXCEED", "Limit should not exceed 10000 members")
		}

		bot.Logger.Infoln("Loading chat members. Offset ", offset)
		m := &tdlib.ChatMembers{}
		var filter tdlib.SupergroupMembersFilter

		for ; limit > 0; limit -= int32(len(m.Members)) {
			bot.Logger.Debugln("Left members", limit)

			//filter = tdlib.NewSupergroupMembersFilterRecent()
			m, err = bot.Client.GetSupergroupMembers(c.SupergroupID, filter, offset, limit)
			if err != nil {
				return result, err.(*tdlib.Error)
			}
			//если в ответе 0 members значит телеграм заблокировал запросы
			if len(m.Members) == 0 {
				return result, tdlib.NewError(tdlib.ErrorCodeAborted, "CLIENT_PARSING_ABORTED", "Request GetSupergroupMembers blocked")
			}

			offset += int32(len(m.Members))
			result = append(result, m.Members...)
		}
	} else {

		m, err := bot.Client.SearchChatMembers(chat.ID, "", 200000, nil)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}

		if len(m.Members) == 0 {
			return nil, tdlib.NewError(tdlib.ErrorCodeAborted, "BOT_WRONG_DATA", "No members in the chat")
		}

		if offset > int32(len(m.Members)) {
			return nil, tdlib.NewError(tdlib.ErrorCodeAborted, "BOT_WRONG_DATA", "Offset must be less than the chat members")
		}

		if limit == 0 {
			limit = int32(len(m.Members)) - offset
		}

		if limit > 0 && limit+offset < int32(len(m.Members)) {
			result = m.Members[offset : limit+offset]
		} else {
			result = m.Members[offset:]
		}
	}

	fmt.Println("User count", len(result))
	return result, nil
}

// ImortContact импортировать контакт
// @phone Номер телефона пользователя
// @firstname Имя пользователя
// @lastname Фамилия пользователя
func (bot *Bot) ImportContact(phone, firstname, lastname string) (int32, *tdlib.Error) {

	if phone == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty phone value")
	}

	if firstname == "" {
		firstname = phone
	}

	contact := tdlib.Contact{}
	contact.FirstName = firstname
	contact.LastName = lastname
	contact.PhoneNumber = phone

	result, err := bot.Client.ImportContacts([]tdlib.Contact{contact})

	if err != nil {
		return 0, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", err.Error())
	}

	if result.UserIDs[0] == 0 {
		bot.Logger.Debugf("User %s not exists\n", phone)
		return 0, tdlib.NewError(ErrorCodeUserNotExists, "BOT_USER_NOT_EXISTS", "User not exists")
	}

	bot.Logger.Debugf("Contact %s import success. User ID : %d\n", phone, result.UserIDs[0])
	return result.UserIDs[0], nil
}

// AddContact добавить контакт НЕ РАБОТАЕТ!!!!!
// @phone Номер телефона пользователя
// @firstname Имя пользователя
// @lastname Фамилия пользователя
func (bot *Bot) AddContact(uid int32, firstname, lastname string) (int32, *tdlib.Error) {

	if uid == 0 {
		return 0, tdlib.NewError(ErrorCodeWrongData, "Empty user id", "")
	}

	if firstname == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "Empty first name", "")
	}

	contact := tdlib.Contact{}
	contact.FirstName = firstname
	contact.LastName = lastname
	contact.UserID = uid

	result, err := bot.Client.AddContact(&contact, false)
	if err != nil {
		return 0, err.(*tdlib.Error)
	}

	if err != nil {
		bot.Logger.Error(err)
		return 0, tdlib.NewError(ErrorCodeUserNotExists, "Error", err.Error())
	}

	bot.Logger.Debugf("Пользователь %d успешно добавлен : %#v\n", uid, result)
	return uid, nil
}

// SendMessageByPhone отправить сообщение пользователю по номеру телефона
// предварительно проверяется существование пользователя
// @phone -номертелефона пользователя
// @msg - текст сообщения
// @ontime - максимальное время не активности пользователя, если пользователь был в сети до этого времени то сообщение не отправляется
func (bot *Bot) SendMessageByPhone(phone, msg string, ontime int32) (int64, *tdlib.Error) {

	// вводим новое событие для пары ImportContacts и SendMessage
	update := tdlib.UpdateData{
		"@type": "sendMessageByPhone",
	}
	raw, _ := json.Marshal(update)

	//проверить существование пользователя
	uid, importErr := bot.ImportContact(phone, phone, "")
	// ErrorUserExists пользователь есть в контактах
	if importErr != nil && importErr.Code != ErrorCodeContactDuplicate {
		bot.Logger.Error("Import error : ", importErr)
		// Если ошибка пользователь не существует то публикуем событие sendMessageByPhone как свершившееся
		if importErr.Code == ErrorCodeUserNotExists {
			bot.Client.PublishEvent(tdlib.EventTypeResponse, tdlib.UpdateMsg{Data: update, Raw: raw})
		}
		return 0, importErr
	}

	mid, sendErr := bot.SendMessageByUID(uid, msg, ontime)

	if sendErr == nil {
		bot.Logger.Debugf("Send to %s success. Message ID : %d\n", phone, mid)
	}

	bot.Client.PublishEvent(tdlib.EventTypeResponse, tdlib.UpdateMsg{Data: update, Raw: raw})

	return mid, sendErr
}

// SendMessageByName отправить сообщение пользователю по его @username
// @username - имя получателя
// @ontime - максимальное время не активности пользователя, если пользователь был в сети до этого времени то сообщение не отправляется
func (bot *Bot) SendMessageByName(username string, msg string, ontime int32) (int64, *tdlib.Error) {

	if msg == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty message value")
	}

	chats, err := bot.Client.SearchPublicChats(username)

	if err != nil {
		return 0, err.(*tdlib.Error)
	}

	if len(chats.ChatIDs) == 0 {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "User not found")
	}

	mid, sendErr := bot.SendMessageByUID(int32(chats.ChatIDs[0]), msg, ontime)

	if sendErr == nil {
		bot.Logger.Debugf("Send to %s success. Message ID : %d\n", username, mid)
	}

	return mid, sendErr
}

// SendMessageByID отправить сообщение пользователю по его ID.
// @ontime - максимальное время не активности пользователя, если пользователь был в сети до этого времени то сообщение не отправляется
func (bot *Bot) SendMessageByUID(uid int32, msg string, ontime int32) (int64, *tdlib.Error) {

	if uid == 0 {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty user ID value")
	}

	if msg == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty message value")
	}

	usr, err := bot.Client.GetUser(uid)
	if err != nil {
		return 0, err.(*tdlib.Error)
	}

	//если установлен ontime проверить время последней активности пользователя
	if ontime != 0 {
		if usr.Status.GetUserStatusEnum() == tdlib.UserStatusOfflineType {
			//"userStatusOffline" {
			s := usr.Status.(*tdlib.UserStatusOffline)
			need := int32(time.Now().Unix()) - ontime
			if need > s.WasOnline {
				return 0, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "Old user")
			}
		}
	}

	chat, _ := bot.CreateChat(uid)

	if chat == nil {
		return 0, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "Create Chat Error")
	}

	/*
		var chat *tdlib.Chat
		//проверяем существует ли чат с пользователем
		bot.GetChatList(200)
		chats, err := bot.Client.SearchChats(usr.FirstName, 100)

		if err != nil {
			return 0, err.(*tdlib.Error)
		}

		//Если чат существует то загружаем его, иначе создаем новый
		if chats.TotalCount > 0 {
			chat, _ = bot.Client.GetChat(chats.ChatIDs[0])
		} else {
			chat, _ = bot.CreateChat(uid)
		}

		if chat == nil {
			return 0, tdlib.NewError(ErrorSystem, "Create Chat Error", "")
		}
	*/
	//bot.Logger.Infof("CHAT %#v\n", chat)
	return bot.SendMessageByCID(chat.ID, msg)

}

// SendMessageByCID отправить сообщение в чат по его ID
func (bot *Bot) SendMessageByCID(cid int64, msg string) (int64, *tdlib.Error) {

	ft, _ := bot.Client.ParseMarkdown(tdlib.NewFormattedText(msg, nil))
	inputMsgTxt := tdlib.NewInputMessageText(ft, true, false)

	m, err := bot.Client.SendMessage(cid, 0, 0, nil, nil, inputMsgTxt)

	if err != nil {
		bot.Logger.Errorf("Send Message ERROR :%#v\n", err)
		return 0, err.(*tdlib.Error)
	}

	e := bot.WaitMessageState(m)
	if e != nil {
		return 0, e
	}

	return m.ID, nil
}

// Дождаться статус отправки сообщения
func (bot *Bot) WaitMessageState(msg *tdlib.Message) *tdlib.Error {
	//проверяем статус сообщения. Проверить на большом кол-ве сообщений в чате.
	//Переделать!!!(что бы не загружал список всех сообщений. А еще лучше сделать асинхронное получение статуса)
	for {
		time.Sleep(500 * time.Millisecond)

		mm, err := bot.Client.GetChatHistory(msg.ChatID, 0, 0, 1, true)
		if err != nil {
			bot.Logger.Errorf("Send Message ERROR :%#v\n", err)
			return err.(*tdlib.Error)
		}

		if mm == nil {
			return tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "Failed to get chat history")
		}

		m := &mm.Messages[0]
		if m.SendingState == nil {
			break
		}

		//bot.Logger.Infof("State : %s\n", m.SendingState.GetMessageSendingStateEnum())

		if m.SendingState.GetMessageSendingStateEnum() == tdlib.MessageSendingStateFailedType {
			//"messageSendingStateFailed" {
			//bot.Logger.Errorln("Send Message ERROR : ", m.SendingState.GetMessageSendingStateEnum())

			//переносим аккаун в директорию unauthorized, она указывается относительно директории с профилем
			//err := bot.Profile.Move(bot.Profile.BaseDir() + "spam")
			err := bot.ProfileToSpam()
			if err != nil {
				bot.Logger.Error(err)
				os.Exit(1)
			}

			go bot.Restart()

			return tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "Send Message ERROR : MessageSendingStateFailed")
		}
		//иначе сообщение в процессе отправки. Ждем 1 сек

	}

	return nil
}

//Загрузить информацию о чате
// @chatID - идентификатор чата, может быть имя чата, ссылка на него, chat id
func (bot *Bot) GetChat(chatID string, join bool) (*tdlib.Chat, *tdlib.Error) {

	if chatID == "" {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty chat ID")
	}

	var chat *tdlib.Chat
	var err error

	inviteMarker := "/joinchat/"
	//Если пригласительная ссылка
	if strings.Index(chatID, inviteMarker) >= 0 {
		var chatInfo *tdlib.ChatInviteLinkInfo
		chatInfo, err = bot.Client.CheckChatInviteLink(chatID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}

		//Если уже в группе
		if chatInfo.ChatID != 0 {
			bot.Logger.Debugln("This bot is already in the group")
			chat, _ := bot.Client.GetChat(chatInfo.ChatID)
			return chat, nil
		}

		//Вступаем в группу
		chat, err = bot.Client.JoinChatByInviteLink(chatID)
		if err != nil {
			bot.Logger.Errorf("Join to %s error: %s\n", chatID, err)
			return nil, err.(*tdlib.Error)
		}

	} else {

		//Если chatID == int64 тогда ищем чат по id иначе по имени
		cid, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			chat, err = bot.Client.SearchPublicChat(chatID)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
		} else {
			chat, err = bot.Client.GetChat(cid)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
		}

		if join {
			_, err := bot.Client.JoinChat(chat.ID)
			if err != nil {
				bot.Logger.Errorf("Join to %s error: %s\n", chat.ID, err)
				return nil, err.(*tdlib.Error)
			}
		}
		/*
			if chat.Type.GetChatTypeEnum() == tdlib.ChatTypeSupergroupType {
				c := chat.Type.(*tdlib.ChatTypeSupergroup)
				inf, _ := bot.Client.GetSupergroup(c.SupergroupID)
				fmt.Printf("INFO %#v\n", inf)
			}
		*/
	}

	bot.Logger.Debugf("Get Chat %s success\n", chat.Title)

	time.Sleep(time.Millisecond * 500)
	/*
		chat, err = bot.Client.GetChat(chat.ID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
	*/
	return chat, nil

}

//GetChatList загрузить список чатов
func (bot *Bot) GetChatList(limit int) ([]*tdlib.Chat, error) {
	var allChats []*tdlib.Chat
	var haveFullChatList bool
	err := bot.getChatList(limit, &allChats, &haveFullChatList)

	return allChats, err
}

func (bot *Bot) getChatList(limit int, allChats *[]*tdlib.Chat, haveFullChatList *bool) error {

	if !*haveFullChatList && limit > len(*allChats) {
		offsetOrder := int64(math.MaxInt64)
		offsetChatID := int64(0)
		var chatList = tdlib.NewChatListMain()
		var lastChat *tdlib.Chat

		if len(*allChats) > 0 {
			tChats := *allChats
			lastChat = tChats[len(tChats)-1]
			for i := 0; i < len(lastChat.Positions); i++ {
				//Find the main chat list
				if lastChat.Positions[i].List.GetChatListEnum() == tdlib.ChatListMainType {
					offsetOrder = int64(lastChat.Positions[i].Order)
				}
			}
			offsetChatID = lastChat.ID
		}

		// get chats (ids) from tdlib
		chats, err := bot.Client.GetChats(chatList, tdlib.JSONInt64(offsetOrder),
			offsetChatID, int32(limit-len(*allChats)))
		if err != nil {
			return err
		}
		if len(chats.ChatIDs) == 0 {
			*haveFullChatList = true
			return nil
		}

		for _, chatID := range chats.ChatIDs {
			// get chat info from tdlib
			chat, err := bot.Client.GetChat(chatID)
			if err == nil {
				*allChats = append(*allChats, chat)
			} else {
				return err
			}
		}
		return bot.getChatList(limit, allChats, haveFullChatList)
	}
	return nil
}

// CreateChat создать новый чат с пользователем
func (bot *Bot) CreateChat(uid int32) (*tdlib.Chat, *tdlib.Error) {
	/*
		//Secret chat
		chat, err := bot.Client.CreateNewSecretChat(uid)
		time.Sleep(time.Second * 3)
	*/

	chat, err := bot.Client.CreatePrivateChat(uid, true)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	/*
		var chat *tdlib.Chat
		//проверяем существует ли чат с пользователем
		bot.GetChatList(200)
		chats, err := bot.Client.SearchChats(usr.FirstName, 100)

		if err != nil {
			return 0, err.(*tdlib.Error)
		}

		//Если чат существует то загружаем его, иначе создаем новый
		if chats.TotalCount > 0 {
			chat, _ = bot.Client.GetChat(chats.ChatIDs[0])
		} else {
			chat, _ = bot.CreateChat(uid)
		}

		if chat == nil {
			return 0, tdlib.NewError(ErrorSystem, "Create Chat Error", "")
		}

	*/

	//fmt.Printf("EEEEEE %#v\n", chat)

	return chat, nil
}
