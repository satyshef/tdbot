package tdbot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tdc "github.com/satyshef/go-tdlib/client"
	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/chat"
)

const checkMemberTimeout = 2000 //millisec

// ============================ NEW METHODS ======================================================
// Собрать все не прочитаные сообщения. Сообщение загружаются со всех чатов в которых состоит бот
func (bot *Bot) GetNewMessagesAll(chatLimit int32) ([]tdlib.Message, *tdlib.Error) {
	var result []tdlib.Message
	chats, err := bot.GetChatList(chatLimit)
	if err != nil {
		return nil, err
	} else {
		for _, c := range chats {
			for c.UnreadCount != 0 {
				msgs, err := bot.Client.GetChatHistory(c.ID, c.LastReadInboxMessageID, -99, 99, false)
				history := msgs.Messages[:len(msgs.Messages)-1]
				if err != nil {
					bot.Logger.Errorln("Get chat history : ", err)
					break
				} else {
					var ids []int64
					// Помечаем сообщения как прочитаные
					for _, m := range history {
						var senderID int64
						switch m.Sender.GetMessageSenderEnum() {
						case tdlib.MessageSenderChatType:
							senderID = m.Sender.(*tdlib.MessageSenderChat).ChatID
						case tdlib.MessageSenderUserType:
							senderID = m.Sender.(*tdlib.MessageSenderUser).UserID
						}
						if senderID == bot.Profile.User.ID {
							continue
						}
						result = append(result, m)
						ids = append(ids, m.ID)
					}
					_, err := bot.Client.ViewMessages(c.ID, 0, ids, true)
					if err != nil {
						bot.Logger.Errorln(err)
					}
					c, err = bot.Client.GetChat(c.ID)
					if err != nil {
						bot.Logger.Errorln(err)
						time.Sleep(time.Second * 1)
					}
				}
			}
			time.Sleep(time.Second * 2)
		}

	}

	return result, nil
}

// Собрать все не прочитаные сообщения. Сообщение загружаются со всех чатов в которых состоит бот
// @chatLimit - максимальное количество обрабатываемых чатов
// @msgLimit - максималькое количество сообщений. Максимум 100
func (bot *Bot) GetNewMessages(chatLimit int, msgLimit int32) ([]tdlib.Message, *tdlib.Error) {
	var result []tdlib.Message
	var chatCount int
	chats, err := bot.GetChatList(500)
	//fmt.Println(len(chats))
	//os.Exit(1)
	if err != nil {
		return nil, err
	} else {
		for _, c := range chats {
			if c.UnreadCount == 0 {
				continue
			}
			msgs, err := bot.Client.GetChatHistory(c.ID, c.LastReadInboxMessageID, -msgLimit, msgLimit, false)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
			if len(msgs.Messages) == 0 {
				continue
			}
			history := msgs.Messages[:len(msgs.Messages)-1]
			if err != nil {
				bot.Logger.Errorln("Get chat history : ", err)
				break
			} else {
				var ids []int64
				// Помечаем сообщения как прочитаные
				for _, m := range history {
					var senderID int64
					switch m.Sender.GetMessageSenderEnum() {
					case tdlib.MessageSenderChatType:
						senderID = m.Sender.(*tdlib.MessageSenderChat).ChatID
					case tdlib.MessageSenderUserType:
						senderID = m.Sender.(*tdlib.MessageSenderUser).UserID
					}
					if senderID == bot.Profile.User.ID {
						continue
					}
					result = append(result, m)
					ids = append(ids, m.ID)
				}
				_, err := bot.Client.ViewMessages(c.ID, 0, ids, true)
				if err != nil {
					bot.Logger.Errorln(err)
				}
				if err != nil {
					bot.Logger.Errorln(err)
				}
				chatCount++
			}

			if chatCount >= chatLimit {
				break
			}
		}

	}

	return result, nil
}

// Прочитать сообщения из чата
// @chat - целевой чат
// @msgLimit - максималькое количество сообщений. Максимум 100
func (bot *Bot) GetLastMessage(chat *tdlib.Chat) (*tdlib.Message, *tdlib.Error) {
	//var result []tdlib.Message
	if chat == nil || chat.LastMessage == nil {
		return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "No last message ID")
	}
	msgs, err := bot.Client.GetChatHistory(chat.ID, chat.LastMessage.ID, -1, 1, false)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	if len(msgs.Messages) == 0 {
		return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "No message")
	}
	msg := msgs.Messages[0]

	//fmt.Println("LEN", len(msgs.Messages))
	//history := msgs.Messages[:len(msgs.Messages)-1]
	/*
		var ids []int64
		// Помечаем сообщения как прочитаные
		for _, m := range msgs.Messages {
			var senderID int64
			switch m.Sender.GetMessageSenderEnum() {
			case tdlib.MessageSenderChatType:
				senderID = m.Sender.(*tdlib.MessageSenderChat).ChatID
			case tdlib.MessageSenderUserType:
				senderID = m.Sender.(*tdlib.MessageSenderUser).UserID
			}
			if senderID == bot.Profile.User.ID {
				continue
			}
			result = append(result, m)
			ids = append(ids, m.ID)
		}
	*/
	// Помечаем сообщения как прочитаные
	_, err = bot.Client.ViewMessages(chat.ID, 0, []int64{msg.ID}, true)
	if err != nil {
		bot.Logger.Errorln(err)
	}

	return &msg, nil
}

// SendMessageToGroup отправить сообщение в чат
// @name - имя группы
// @msg - текст сообщения
func (bot *Bot) SendMessageToChat(cid string, msg string, leave bool) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if msg == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty message content")
	}

	chat, err := bot.GetChat(cid, true)
	if err != nil {
		return 0, err
	}
	//bot.Logger.Debugf("Chat %#v\n", chat)
	/*
		ft, _ := bot.Client.ParseMarkdown(tdlib.NewFormattedText(msg, nil))
		inputMsgTxt := tdlib.NewInputMessageText(ft, true, false)
		mm, e := bot.Client.SendMessage(chat.ID, 0, 0, nil, nil, inputMsgTxt)
		fmt.Printf("MES %#v\n\n", mm)
		if e != nil {
			return e.(*tdlib.Error)
		}
	*/

	mid, err := bot.SendMessageByCID(chat.ID, msg)
	if leave {
		bot.Client.LeaveChat(chat.ID)
	}
	return mid, err
	//bot.Logger.Debugf("Send to %s success\n", name)
	//return nil
}

// SendMessageByCID отправить сообщение в чат по его ID
// @cid - ID чата
// @msg - текст сообщения
func (bot *Bot) SendMessageByCID(cid int64, msg string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
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

// CreatePrivateChat создать новый чат с пользователем
// @uid - ID пользователя с которым нужно создать чат
func (bot *Bot) CreatePrivateChat(uid int64) (*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chat, err := bot.Client.CreatePrivateChat(uid, true)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return chat, nil
}

// CreateChannel создать новый канал
// @address
// @name
// @description
func (bot *Bot) CreateChannel(name, address, description string) (*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Если указан адрес тогда проверяем существование такого чата
	if address != "" {
		chat, err := bot.GetChat(address, false)
		if err != nil && err.Code != tdc.ErrorCodeUsernameNotOccupied {
			return nil, err

		}
		if chat != nil {
			return nil, tdlib.NewError(ErrorCodeUserExists, "USERNAME_IS_EXISTS", "Username is exists")
		}
	}
	chat, err := bot.Client.CreateNewSupergroupChat(name, true, description, nil, false)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	if address != "" {
		_, err = bot.Client.SetSupergroupUsername(chat.Type.(*tdlib.ChatTypeSupergroup).SupergroupID, address)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
	}
	return chat, nil
}

// ============================ OLD METHODS ======================================================

// Глобальный поиск чатов
// @query - ключевая фраза поиска
func (bot *Bot) SearchChats(query string) ([]*chat.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chats, err := bot.Client.SearchPublicChats(query)

	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	bot.Logger.Debugf("Find : %s - %d", query, chats.TotalCount)

	result := make([]*chat.Chat, 0)

	for _, cid := range chats.ChatIDs {

		//Загружаем полную информацию о чате
		full, err := bot.GetChatFullInfo(cid)
		if err != nil {
			continue
		}

		if full != nil {
			//bot.Logger.Debugf("FULL CHAT :\n%#v\n\n", full)
			result = append(result, full)
		}
	}

	return result, nil
}

/*
//groups, e := bot.Client.GetChats(nil, 0, 0, 1000)
//Список чатов в которых состоит бот
func (bot *Bot) GetChats() ([]*chat.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chats, err := bot.Client.GetChats(nil, 0, 0, 1000)

	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	bot.Logger.Debugf("The bot consists of %d chats", chats.TotalCount)

	result := make([]*chat.Chat, 0)

	for _, cid := range chats.ChatIDs {

		//Загружаем полную информацию о чате
		full, err := bot.GetChatFullInfo(cid)
		if err != nil {
			continue
		}

		if full != nil {
			//bot.Logger.Debugf("FULL CHAT :\n%#v\n\n", full)
			result = append(result, full)
		}
	}

	return result, nil
}
*/
// Получить полную информацию о чате
// @cid - чат id
func (bot *Bot) GetChatFullInfo(cid int64) (*chat.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chatInfo, err := bot.Client.GetChat(cid)

	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	switch chatInfo.Type.GetChatTypeEnum() {
	case tdlib.ChatTypeBasicGroupType:
		group, err := bot.Client.GetBasicGroup(chatInfo.Type.(*tdlib.ChatTypeBasicGroup).BasicGroupID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}

		chat := chat.New(chatInfo.ID, chatInfo.Title, "", chat.TypeGroup)
		chat.DateCreation = 0
		chat.HasLinkedChat = false
		chat.IsScam = false
		chat.IsVerified = false
		chat.MemberCount = group.MemberCount

		//Не корректно отображает дату последнего сообщения
		if chatInfo.LastMessage != nil {
			chat.DateLastMessage = chatInfo.LastMessage.Date
		}

		// get caht description
		f, err := bot.Client.GetBasicGroupFullInfo(group.ID)
		if err == nil {
			chat.BIO = f.Description
		}

		return chat, nil

	case tdlib.ChatTypeSupergroupType:
		//supergroup
		var chatType chat.Type
		superGroup, err := bot.Client.GetSupergroup(chatInfo.Type.(*tdlib.ChatTypeSupergroup).SupergroupID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		if superGroup.IsChannel {
			chatType = chat.TypeChannel
		} else {
			chatType = chat.TypeGroup
		}

		//bot.Logger.Infof("FULL CHAT :\n%#v\n\n", superGroup)
		//bot.Logger.Infof("Sender :\n%#v\n\n", chatInfo.LastMessage.Date)

		chat := chat.New(chatInfo.ID, chatInfo.Title, superGroup.Usernames.EditableUsername, chatType)
		chat.DateCreation = superGroup.Date
		chat.HasLinkedChat = superGroup.HasLinkedChat
		chat.IsScam = superGroup.IsScam
		chat.IsVerified = superGroup.IsVerified
		chat.MemberCount = superGroup.MemberCount

		//Не корректно отображает дату последнего сообщения
		if chatInfo.LastMessage != nil {
			chat.DateLastMessage = chatInfo.LastMessage.Date
		}

		// get caht description
		f, err := bot.Client.GetSupergroupFullInfo(superGroup.ID)
		if err == nil {
			chat.BIO = f.Description
		}

		return chat, nil

	case tdlib.ChatTypePrivateType:
		//user action...
	default:
		bot.Logger.Infof("UNKNOWN CHAT TYPE:\n\n%#v\n\n", chatInfo)
	}

	return nil, nil
}

func (bot *Bot) InviteByUserName(username, chatname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Блокируем остановку бота до завершения работы функции
	bot.wg.Add(1)
	defer bot.wg.Done()

	destChat, e := bot.GetChat(chatname, true)
	if e != nil {
		return 0, e
	}
	userChat, err := bot.Client.SearchPublicChat(username)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
		return 0, e
	}
	_, err = bot.Client.AddChatMember(destChat.ID, userChat.ID, 100)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
		return 0, e
	} else {
		bot.Logger.Debugf("Invite user %s - OK", username)
	}
	// Проверяем добавился ли пользователь
	e = bot.CheckMember(destChat.ID, userChat.ID, checkMemberTimeout)
	if e != nil {
		return 0, e
	}
	return userChat.ID, nil
}

func (bot *Bot) AddAdminChat(username, chatname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Блокируем остановку бота до завершения работы функции
	bot.wg.Add(1)
	defer bot.wg.Done()

	destChat, e := bot.GetChat(chatname, true)
	if e != nil {
		return 0, e
	}
	userChat, err := bot.Client.SearchPublicChat(username)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Add admin %s - %s", username, e.Message)
		return 0, e
	}

	//status := tdlib.NewChatMemberStatusAdministrator("", true, false, false, false, false, false, true, false, false, false, false, false)
	status := tdlib.NewChatMemberStatusAdministrator("Admin", true, true, false, false, false, false, false, false, false, false, false, false, false)
	//fmt.Printf("STATUS %#v\n\n", status)
	//st := tdlib.NewChatMemberStatusMember()
	member := tdlib.NewMessageSenderUser(userChat.ID)
	_, err = bot.Client.SetChatMemberStatus(destChat.ID, member, status)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Add admin %s - %s", username, e.Message)
		return 0, e
	} else {
		bot.Logger.Debugf("Add admin %s - OK", username)
	}
	// Проверяем добавился ли пользователь
	e = bot.CheckMember(destChat.ID, userChat.ID, checkMemberTimeout)
	if e != nil {
		return 0, e
	}
	return userChat.ID, nil
}

func (bot *Bot) InviteByUserNameTest(username, chatname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Блокируем остановку бота до завершения работы функции
	bot.wg.Add(1)
	defer bot.wg.Done()

	users, err := bot.Client.SearchContacts(username, 10)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
		return 0, e
	}

	var userID int64

	if users.TotalCount == 0 {
		userChat, err := bot.Client.SearchPublicChat(username)
		if err != nil {
			e := err.(*tdlib.Error)
			bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
			return 0, e
		}
		_, e := bot.AddContact(userChat.ID, username, "")
		if e != nil {
			return 0, e
		}
		userID = userChat.ID
	} else {
		userID = users.UserIDs[0]
	}

	destChat, e := bot.GetChat(chatname, true)
	if e != nil {
		return 0, e
	}

	_, err = bot.Client.AddChatMember(destChat.ID, userID, 100)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
		return 0, e
	} else {
		bot.Logger.Debugf("Invite user %s - OK", username)
	}

	return userID, nil
}

// инвайт по имени пользователя
func (bot *Bot) InviteByName(name string, userID int64, chatNameSource, chatNameDest string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Блокируем остановку бота до завершения работы функции
	bot.wg.Add(1)
	defer bot.wg.Done()

	sourceChat, e := bot.GetChat(chatNameSource, true)
	if e != nil {
		return 0, e
	}
	var filter tdlib.ChatMembersFilter
	members, err := bot.Client.SearchChatMembers(sourceChat.ID, name, 200, filter)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Search user %s - %s", name, e.Message)
		return 0, e
	}
	var uid int64
	for _, m := range members.Members {

		if m.MemberID.(*tdlib.MessageSenderUser).UserID == userID {
			uid = userID
			break
		}
	}

	if uid == 0 {
		bot.Logger.Debugf("User %s not found", name)
		return 0, tdlib.NewError(tdc.ErrorCodeMemberNotFound, "CLIENT_MEMBER_NOT_FOUND", "")
	}

	//TODO: test. added user to contact
	/*
		_, err = bot.Client.SearchContacts(name, 10)
		if err != nil {
			e := err.(*tdlib.Error)
			bot.Logger.Debugf("Invite user %s - %s", name, e.Message)
			return 0, e
		}
	*/
	_, e = bot.AddContact(uid, name, "")
	if e != nil {
		return 0, e
	}

	destChat, e := bot.GetChat(chatNameDest, true)
	if e != nil {
		return 0, e
	}

	_, err = bot.Client.AddChatMember(destChat.ID, uid, 100)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", name, e.Message)
		return 0, e
	}

	// Проверяем добавился ли пользователь
	e = bot.CheckMember(destChat.ID, uid, checkMemberTimeout)
	if e != nil {
		return 0, e
	}
	bot.Logger.Debugf("Invite user %s - OK", name)
	return uid, nil
}

func (bot *Bot) InviteByUserPhone(phone, chatname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Блокируем остановку бота до завершения работы функции
	bot.wg.Add(1)
	defer bot.wg.Done()

	destChat, e := bot.GetChat(chatname, true)
	if e != nil {
		return 0, e
	}
	//bot.Logger.Debugf("Add phone %s to contacts", phone)

	//проверить существование пользователя
	uid, e := bot.ImportContact(phone, phone, "")
	// ErrorUserExists пользователь есть в контактах
	if e != nil && e.Code != ErrorCodeContactDuplicate {
		return 0, e
	}

	_, err := bot.Client.AddChatMember(destChat.ID, uid, 100)
	if err != nil {
		bot.Logger.Error("%#v\n", err)
		return 0, err.(*tdlib.Error)
	} else {
		bot.Logger.Infoln("Add OK")
	}

	// Проверяем добавился ли пользователь
	e = bot.CheckMember(destChat.ID, uid, checkMemberTimeout)
	if e != nil {
		return 0, e
	}

	return uid, nil
}

// TODO : работает криво. прежде чем добавить нужно получить участника методом GetChatMembers
func (bot *Bot) InviteByID(uid, cid int64) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// Блокируем остановку бота до завершения работы функции
	bot.wg.Add(1)
	defer bot.wg.Done()
	/*
		destChat, e := bot.GetChat(chat, false)
		if e != nil {
			return e
		}
	*/
	//bot.Logger.Debugln("Add user : ", uid)
	/*
		userChat, err := bot.Client.SearchPublicChat(userchat)
		if err != nil {
			return err.(*tdlib.Error)
		}
	*/

	_, err := bot.Client.AddChatMember(cid, uid, 100)
	if err != nil {
		//bot.Logger.Errorf("%#v\n", err)
		return err.(*tdlib.Error)
	}
	// Проверяем добавился ли пользователь
	e := bot.CheckMember(cid, uid, checkMemberTimeout)
	if e != nil {
		return e
	}
	bot.Logger.Debugf("Invite user %d - %s", uid, "OK")
	//bot.Logger.Info(" - OK")
	return nil
}

// Проверяем добавился ли пользователь в группу
func (bot *Bot) CheckMember(cid, uid, timeout int64) *tdlib.Error {
	time.Sleep(time.Millisecond * time.Duration(timeout))
	member := tdlib.NewMessageSenderUser(uid)
	m, err := bot.Client.GetChatMember(cid, member)
	if err != nil {
		//bot.Logger.Errorf("%#v\n", err)
		return err.(*tdlib.Error)
	}
	if m.Status.GetChatMemberStatusEnum() == tdlib.ChatMemberStatusLeftType {
		// вводим новое событие
		ev := tdc.NewEvent(tdc.EventTypeResponse, EventNameChatMemberLeft, 0, "")
		bot.Client.PublishEvent(ev)
		return tdlib.NewError(ErrorCodeChatMemberLeft, "CHAT_MEMBER_LEFT", "Telegram auto remove member from chat")
	}

	return nil
}

//Получить полную информацию о пользователе. Переделать что бы брать инфу по username!!!
// @uid - ID пользователя
// @chatName - имя чата в которой находится пользователь
func (bot *Bot) GetUser(userID string) (*tdlib.User, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}

	chat, errTd := bot.GetChat(userID, false)
	if errTd != nil {
		return nil, errTd
	}

	//TODO: эксперемент. Не обязательный запрос
	/*
		_, e := bot.AddContact(chat.ID, userID, "")
		if e != nil {
			return nil, e
		}
	*/

	u, err := bot.Client.GetUser(chat.ID)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return u, nil
}

//Получить полную информацию о пользователе по его номеру телефона.
// @uid - ID пользователя
// @chatName - имя чата в которой находится пользователь
func (bot *Bot) GetUserByPhone(phone string) (*tdlib.User, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	//проверить существование пользователя
	userID, importErr := bot.ImportContact(phone, phone, "")
	// ErrorUserExists пользователь есть в контактах
	if importErr != nil && importErr.Code != ErrorCodeContactDuplicate {
		return nil, importErr
	}
	u, err := bot.Client.GetUser(userID)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return u, nil
}

// Отправить сообщение участникам группы
func (bot *Bot) SendMessageChatMembers(chatAddr string, message string, offset int32, limit int32) ([]tdlib.ChatMember, *tdlib.Error) {

	result := []tdlib.ChatMember{}
	members, err := bot.GetChatMembers(chatAddr, offset, limit, true)

	if len(members) > 0 {
		for _, m := range members {
			uid := m.MemberID.(*tdlib.MessageSenderUser).UserID
			_, err := bot.SendMessageByUID(uid, message, 0)
			time.Sleep(time.Second * 2)
			if err != nil {
				if err.Code == tdc.ErrorCodeNoAccess {
					bot.Logger.Errorf("%d - %s", uid, err.Message)
					continue
				}
				return result, err
			}
			result = append(result, m)
			bot.Logger.Printf("[ %s ] Send to user ID %d - OK\n", time.Now().Format(time.RFC3339), uid)

		}
	}
	return result, err

}

/*
// Загрузить полную информацию об участниках группы
func (bot *Bot) GetChatUsers_(chatname string, offset int32, limit int32) ([]tdlib.User, *tdlib.Error) {
	if limit < 0 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Limit must be positive")
	}
	if offset < 0 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Offset must be positive")
	}
	result := []tdlib.User{}
	members, err := bot.GetChatMembers(chatname, offset, limit)
	if len(members) > 0 {
		for _, m := range members {
			user, err := bot.Client.GetUser(m.MemberID.GetID())
			if err != nil {
				bot.Logger.Error("Get User", err)
				fmt.Printf("%#v\n\n", result)
				break
			} else {
				result = append(result, *user)
			}
		}
	}
	return result, err
}
*/

// Загрузить список пользователей группы с полной информацией о них
func (bot *Bot) GetChatUsers(chatname string, offset int32, limit int32, join bool) (result []tdlib.User, err *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if limit < 0 || limit > 10000 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Limit must be positive and less than 10000")
	}
	if offset < 0 || offset > 10000 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Offset must be positive and less than 10000")
	}
	if limit == 0 {
		limit = 10000
	}
	if (offset + limit) > 10000 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Sum offset and limit should not exceed 10000")
		//limit = 10000 - offset
	}

	var members []tdlib.ChatMember
	var count int32 = 200

	if limit < count {
		count = limit
	}

	//return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Sum offset and limit should not exceed 10000")

	for ; limit > 0; limit -= count {
		members, err = bot.GetChatMembers(chatname, offset, count, join)
		len := int32(len(members))
		if err != nil || len == 0 {
			break
		}

		for _, m := range members {
			uid := m.MemberID.(*tdlib.MessageSenderUser).UserID
			user, e := bot.Client.GetUser(uid)
			if e != nil {
				err = e.(*tdlib.Error)
				bot.Logger.Error("Get User", err)
				break
			} else {
				result = append(result, *user)
			}
		}

		//если количество упользователей меньше count то выходим из цикла

		//	ПЕРЕДЕЛАТЬ!!!!!
		if count > len {
			break
		}

		offset += len

	}
	return result, err
}

// Загрузить список участников группы. Группа может быть как обычная(до 200) так и супергруппа(до 200к) участников. Возвращает ID участников и их статус
// @chatname - имя или сслыка на чат
func (bot *Bot) GetChatMembers(chatname string, offset int32, limit int32, join bool) ([]tdlib.ChatMember, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if limit < 0 || limit > 10000 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Limit must be positive and less than 10000")
	}
	if offset < 0 || offset > 10000 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Offset must be positive and less than 10000")
	}
	var chat *tdlib.Chat
	var err error
	var result []tdlib.ChatMember
	//TODO: плохая идея оставаться в чате
	chat, e := bot.GetChat(chatname, join)
	if e != nil {
		return nil, e
	}
	bot.Logger.Debugln("Parsing chat : ", chatname)

	//return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Offset must be positive")
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
			limit = 10000
		}
		bot.Logger.Debugln("Loading chat members. Offset ", offset)
		m := &tdlib.ChatMembers{}
		var filter tdlib.SupergroupMembersFilter
		var count int32 = 200
		if limit < count {
			count = limit
		}
		for ; limit > 0; limit -= count {
			bot.Logger.Debugln("Left members", limit)
			m, err = bot.Client.GetSupergroupMembers(c.SupergroupID, filter, offset, count)
			if err != nil {
				return result, err.(*tdlib.Error)
			}
			len := int32(len(m.Members))
			//если в ответе 0 members значит телеграм заблокировал запросы
			if len == 0 {
				return result, tdlib.NewError(tdc.ErrorCodeAborted, "CLIENT_PARSING_ABORTED", "Request GetSupergroupMembers blocked")
			}

			result = append(result, m.Members...)
			//если количество загруженных пользователей меньше count то выходим из цикла
			if count > len {
				break
			}

			offset += len
		}
	} else {
		m, err := bot.Client.SearchChatMembers(chat.ID, "", 200000, nil)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		if len(m.Members) == 0 {
			return nil, tdlib.NewError(tdc.ErrorCodeAborted, "BOT_WRONG_DATA", "No members in the chat")
		}
		if offset > int32(len(m.Members)) {
			return nil, tdlib.NewError(tdc.ErrorCodeAborted, "BOT_WRONG_DATA", "Offset must be less than the chat members")
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

// Скопировать участников одной группы в другую
func (bot *Bot) CopyGroupMembers(source, destination string, offset int32, limit int32) ([]tdlib.ChatMember, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	result := []tdlib.ChatMember{}

	// Загружаем список участников группы-донора
	members, err := bot.GetChatMembers(source, offset, limit, true)
	if len(members) > 0 {
		destChat, e := bot.GetChat(destination, true)
		if e != nil {
			return nil, e
		}

		// Добавляем участников в целевую группу
		for _, m := range members {
			uid := m.MemberID.(*tdlib.MessageSenderUser).UserID
			// Исключаем себя
			if bot.Profile.User.ID == uid {
				continue
			}

			user, err := bot.Client.GetUser(uid)
			if err != nil {
				bot.Logger.Errorln("Get User Error ", err)
				continue
			}

			//Фильтр только живые люди
			// Пропускаем скамеров, ботов, администраторов
			if user.IsScam || user.Type.GetUserTypeEnum() == tdlib.NewUserTypeDeleted().GetUserTypeEnum() || user.Type.GetUserTypeEnum() == tdlib.NewUserTypeBot(false, false, false, "", false).GetUserTypeEnum() {
				continue
			}

			bot.Logger.Infoln("Add user ID", uid)
			_, err = bot.Client.AddChatMember(destChat.ID, uid, 100)
			if err != nil {
				//Если заблокировали за флуд возвращаем результат
				if err.(*tdlib.Error).Code == tdc.ErrorCodeFloodLock {
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

// ImortContact импортировать контакт
// @phone Номер телефона пользователя
// @firstname Имя пользователя
// @lastname Фамилия пользователя
func (bot *Bot) ImportContact(phone, firstname, lastname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
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
// @uid id пользователя
// @firstname Имя пользователя
// @lastname Фамилия пользователя
func (bot *Bot) AddContact(uid int64, firstname, lastname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
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

	_, err := bot.Client.AddContact(&contact, false)
	if err != nil {
		return 0, err.(*tdlib.Error)
	}

	if err != nil {
		bot.Logger.Error(err)
		return 0, tdlib.NewError(ErrorCodeUserNotExists, "Error", err.Error())
	}

	bot.Logger.Debugf("Пользователь %d успешно добавлен в контакты", uid)
	return uid, nil
}

// SendMessageByPhone отправить сообщение пользователю по номеру телефона
// предварительно проверяется существование пользователя
// @phone -номертелефона пользователя
// @msg - текст сообщения
// @ontime - максимальное время не активности пользователя, если пользователь был в сети до этого времени то сообщение не отправляется
// Возвращает Message ID
func (bot *Bot) SendMessageByPhone(phone, msg string, ontime int32) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	// вводим новое событие для пары ImportContacts и SendMessage
	ev := tdc.NewEvent(tdc.EventTypeResponse, EventNameSendMessageByPhone, 0, "")
	//проверить существование пользователя
	uid, importErr := bot.ImportContact(phone, phone, "")
	// ErrorUserExists пользователь есть в контактах
	if importErr != nil && importErr.Code != ErrorCodeContactDuplicate {
		// Если ошибка пользователь не существует то публикуем событие sendMessageByPhone как свершившееся
		if importErr.Code == ErrorCodeUserNotExists {
			bot.Client.PublishEvent(ev)
		} else {
			bot.Logger.Error("Import error : ", importErr)
		}
		return 0, importErr
	}
	mid, sendErr := bot.SendMessageByUID(uid, msg, ontime)
	if sendErr == nil {
		bot.Logger.Debugf("Send to %s success. Message ID : %d\n", phone, mid)
	}
	bot.Client.PublishEvent(ev)
	return mid, sendErr
}

// SendMessageByName отправить сообщение пользователю по его @username
// @username - имя получателя
// @ontime - максимальное время не активности пользователя, если пользователь был в сети до этого времени то сообщение не отправляется
// Возвращает Message ID
func (bot *Bot) SendMessageByName(username string, msg string, ontime int32) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
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

	mid, sendErr := bot.SendMessageByUID(chats.ChatIDs[0], msg, ontime)

	if sendErr == nil {
		bot.Logger.Debugf("Send to %s success. Message ID : %d\n", username, mid)
	}

	return mid, sendErr
}

// SendMessageByID отправить сообщение пользователю по его ID.
// @uid - ID пользователя
// @ontime - максимальное время не активности пользователя, если пользователь был в сети до этого времени то сообщение не отправляется
func (bot *Bot) SendMessageByUID(uid int64, msg string, ontime int32) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
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

	chat, _ := bot.CreatePrivateChat(uid)

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

		//Заблокирована отправка сообщений
		if m.SendingState.GetMessageSendingStateEnum() == tdlib.MessageSendingStateFailedType {
			e := tdlib.NewError(tdc.ErrorCodeFloodLock, "PEER_FLOOD", "User spam block")
			ev := tdc.ErrorToEvent(e)
			bot.Client.PublishEvent(ev)
			return e

			//ev := tdlib.NewEvent(tdlib.EventTypeError, EventNameBotReady, 0, "")
			/*
				err := bot.ProfileToSpam()
				if err != nil {
					bot.Logger.Error(err)
					os.Exit(1)
				}

				go bot.Restart()

				return tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "Send Message ERROR : MessageSendingStateFailed")
			*/
		}
		//иначе сообщение в процессе отправки. Ждем 1 сек

	}

	return nil
}

//Загрузить информацию о чате
// @chatID - идентификатор чата, может быть имя чата, ссылка на него, chat id
func (bot *Bot) GetChat(chatname string, join bool) (*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if chatname == "" {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty chat ID")
	}
	var chat *tdlib.Chat
	//var imInGroup bool
	var err error
	//Если пригласительная ссылка
	if strings.Contains(chatname, "t.me/joinchat/") || strings.Contains(chatname, "t.me/+") {
		//bot.Logger.Debugln("Parse private chat ", chatname)
		var chatInfo *tdlib.ChatInviteLinkInfo
		chatInfo, err = bot.Client.CheckChatInviteLink(chatname)
		if err != nil {
			return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", err.Error())
		}
		//Если уже в группе
		if chatInfo.ChatID != 0 {
			bot.Logger.Debugln("This bot is already in the group", chatname)
			chat, _ = bot.Client.GetChat(chatInfo.ChatID)

		} else {
			//Вступаем в группу
			chat, err = bot.Client.JoinChatByInviteLink(chatname)
			if err != nil {
				bot.Logger.Errorf("Join to %s error: %s\n", chatname, err)
				return nil, err.(*tdlib.Error)
			}
		}
	} else {

		//Если chatID == int64 тогда ищем чат по id иначе по имени
		cid, err := strconv.ParseInt(chatname, 10, 64)
		if err != nil {
			// Ищем в своих группах
			// TODO: не всегда находит чат
			//bot.Client.SearchChats(profile.RandomString(5), 1)
			chts, err := bot.Client.SearchChats(chatname, 1)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
			//fmt.Printf("%#v", chts)
			if chts.TotalCount > 0 {
				bot.Logger.Debugln("This bot is already in the group", chatname)
				chat, _ = bot.Client.GetChat(chts.ChatIDs[0])
			} else {
				// Ищем публичную группу
				chat, err = bot.Client.SearchPublicChat(chatname)
				if err != nil {
					return nil, err.(*tdlib.Error)
				}
				if join {
					_, err := bot.Client.JoinChat(chat.ID)
					if err != nil {
						bot.Logger.Errorf("Join to %s error: %s\n", chat.ID, err)
						return nil, err.(*tdlib.Error)
					}
				}
			}
		} else {
			//Получаем группу по ID
			chat, err = bot.Client.GetChat(cid)
			if err != nil {
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

	//bot.Logger.Debugf("Get Chat %s success\n", chat.Title)

	time.Sleep(time.Millisecond * 500)
	/*
		chat, err = bot.Client.GetChat(chat.ID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
	*/
	return chat, nil

}

// Список всех созданных пользователем чатов.
func (bot *Bot) GetCreatedChats() ([]*tdlib.Chat, *tdlib.Error) {
	chatType := tdlib.NewPublicChatTypeHasUsername()
	chats, err := bot.Client.GetCreatedPublicChats(chatType)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	var result []*tdlib.Chat
	for _, cid := range chats.ChatIDs {
		chat, err := bot.Client.GetChat(cid)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		result = append(result, chat)
	}
	return result, nil

}

// Список всех созданных пользователем каналов.
// TODO: отображаются только публичные. Сделать что бы подгружались и скрытые
func (bot *Bot) GetCreatedChannels() ([]*tdlib.Chat, *tdlib.Error) {
	chatType := tdlib.NewPublicChatTypeHasUsername()
	chats, err := bot.Client.GetCreatedPublicChats(chatType)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	var result []*tdlib.Chat
	for _, cid := range chats.ChatIDs {
		chat, err := bot.Client.GetChat(cid)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		if chat.Type.GetChatTypeEnum() == tdlib.ChatTypeSupergroupType {
			if chat.Type.(*tdlib.ChatTypeSupergroup).IsChannel {
				result = append(result, chat)
			}
		}
	}
	return result, nil

}

// Список всех созданных пользователем групп.
// TODO: работает только с супергруппами
func (bot *Bot) GetCreatedGroups() ([]*tdlib.Chat, *tdlib.Error) {
	chatType := tdlib.NewPublicChatTypeHasUsername()
	chats, err := bot.Client.GetCreatedPublicChats(chatType)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	var result []*tdlib.Chat
	for _, cid := range chats.ChatIDs {
		chat, err := bot.Client.GetChat(cid)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		if chat.Type.GetChatTypeEnum() == tdlib.ChatTypeSupergroupType {
			if !chat.Type.(*tdlib.ChatTypeSupergroup).IsChannel {
				result = append(result, chat)
			}
		}
	}
	return result, nil

}

//GetChatList загрузить список чатов аккаунта
func (bot *Bot) GetChatList(limit int32) ([]*tdlib.Chat, *tdlib.Error) {
	var allChats []*tdlib.Chat
	var haveFullChatList bool
	err := bot.getChatList(limit, &allChats, &haveFullChatList)

	return allChats, err
}

// TODO: проверить загружает ли все чаты
func (bot *Bot) getChatList(limit int32, allChats *[]*tdlib.Chat, haveFullChatList *bool) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chats, err := bot.Client.GetChats(nil, limit)
	if err != nil {
		return err.(*tdlib.Error)
	}

	for _, chatID := range chats.ChatIDs {
		// get chat info from tdlib
		chat, err := bot.Client.GetChat(chatID)
		if err == nil {
			*allChats = append(*allChats, chat)
		} else {
			return err.(*tdlib.Error)
		}
	}
	return nil
}

/*
func (bot *Bot) getChatList(limit int, allChats *[]*tdlib.Chat, haveFullChatList *bool) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if !*haveFullChatList && limit > len(*allChats) {
		offsetOrder := int64(math.MaxInt64)
		offsetChatID := int64(0)
		//var chatList = tdlib.NewChatListMain()
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
		currentLimit := int32(limit - len(*allChats))
		chats, err := bot.Client.GetChats(nil, tdlib.JSONInt64(offsetOrder), offsetChatID, currentLimit)
		if err != nil {
			return err.(*tdlib.Error)
		}

		for _, chatID := range chats.ChatIDs {
			// get chat info from tdlib
			chat, err := bot.Client.GetChat(chatID)
			if err == nil {
				*allChats = append(*allChats, chat)
			} else {
				return err.(*tdlib.Error)
			}
		}

		if int32(len(chats.ChatIDs)) < currentLimit {
			*haveFullChatList = true
			return nil
		}
		return bot.getChatList(limit, allChats, haveFullChatList)
	}
	return nil
}
*/
