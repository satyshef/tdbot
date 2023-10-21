package tdbot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tdc "github.com/satyshef/go-tdlib/client"
	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/chat"
	"github.com/satyshef/tdbot/profile"
)

const checkMemberTimeout = 2000 //millisec

// ============================ NEW METHODS ======================================================

// Отметить сообщение как прочитанное
// @chatID - id чата
// @messageIDs - список сообщений которые нужно пометить как прочитаные
func (bot *Bot) ViewMessages(chatID int64, messageIDs []int64) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	_, err := bot.client.ViewMessages(chatID, 0, messageIDs, true)
	if err != nil {
		return err.(*tdlib.Error)
	}

	return nil
}

func (bot *Bot) GetChatHistory(chatID int64, fromMessageID int64, offset int32, limit int32, onlyLocal bool) (*tdlib.Messages, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	msgs, err := bot.client.GetChatHistory(chatID, fromMessageID, offset, limit, onlyLocal)
	if err != nil {
		bot.Logger.Errorln("Get chat history : ", err)
		return nil, err.(*tdlib.Error)
	}
	return msgs, nil
}

// Загрузить историю сообщений чата
// @timeout - задержка между запросами в миллисек
// @limit - лимит сообщений, если 0 тогда значение по умолчанию 100k
func (bot *Bot) GetChatHistoryFull(chatID int64, limit int32, fromMessageID int64, timeout int) ([]tdlib.Message, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	var result []tdlib.Message
	step := int32(99)
	var countResult int32

	if limit == 0 {
		limit = 100000
	}

	for ; limit > 0; limit -= countResult {
		time.Sleep(time.Millisecond * time.Duration(timeout))
		msgs, err := bot.client.GetChatHistory(chatID, fromMessageID, 0, step, false)
		if err != nil {
			bot.Logger.Errorln("Get chat history : ", err)
			return nil, err.(*tdlib.Error)
		}

		countResult = int32(len(msgs.Messages))
		fmt.Println("Parse ", countResult)
		if countResult == 0 {
			break
		}

		result = append(result, msgs.Messages...)
		fromMessageID = msgs.Messages[countResult-1].ID
	}
	return result, nil
}

// Собрать все не прочитаные сообщения.
func (bot *Bot) GetUnreadMessagesAll(chat *tdlib.Chat, timestamp int32) ([]tdlib.Message, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	var result []tdlib.Message
	var fromMessageID int64
	lastMessageID := chat.LastReadInboxMessageID
	unreadCount := chat.UnreadCount
	for unreadCount > 0 {
		fmt.Printf("ID %d\n", fromMessageID)
		msgs, err := bot.client.GetChatHistory(chat.ID, fromMessageID, 0, unreadCount, false)
		if err != nil {
			bot.Logger.Errorln("Get chat history: ", err)
			break
		}

		countResult := int32(len(msgs.Messages))
		if countResult == 0 {
			break
		}

		var ids []int64
		for _, m := range msgs.Messages {
			senderID := getSenderID(m.Sender)
			if senderID == bot.Profile.User.ID {
				continue
			}
			ids = append(ids, m.ID)

			if (timestamp != 0 && timestamp > m.Date) || m.ID == lastMessageID {
				break
				/*
					bot.client.ViewMessages(chat.ID, 0, ids, true)
					bot.client.GetChat(chat.ID)
					return result, nil
				*/
			}

			result = append(result, m)
			unreadCount -= 1
		}

		_, err = bot.client.ViewMessages(chat.ID, 0, ids, true)
		if err != nil {
			bot.Logger.Errorln(err)
		}

		/*
			_, e := bot.GetChatFullInfo(chat.ID)
			if e != nil {
				return nil, e
			}
		*/
		fromMessageID = msgs.Messages[countResult-1].ID
	}

	return result, nil
}

func getSenderID(sender tdlib.MessageSender) int64 {
	switch sender.GetMessageSenderEnum() {
	case tdlib.MessageSenderChatType:
		return sender.(*tdlib.MessageSenderChat).ChatID
	case tdlib.MessageSenderUserType:
		return sender.(*tdlib.MessageSenderUser).UserID
	default:
		return 0
	}
}

// TODO: доработать что бы всю историю загружал
// Собрать не прочитаные сообщения. Сообщение загружаются со всех чатов в которых состоит бот
// @chatLimit - максимальное количество обрабатываемых чатов
// @msgLimit - максималькое количество сообщений. Максимум 100
/*
func (bot *Bot) GetUnreadMessages11111(chatLimit int, msgLimit int32) ([]tdlib.Message, *tdlib.Error) {
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

			msgs, err := bot.client.GetChatHistory(c.ID, c.LastReadInboxMessageID, -msgLimit, msgLimit, false)
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
					senderID := getSenderID(m.Sender)
					if senderID == bot.Profile.User.ID {
						continue
					}
					result = append(result, m)
					ids = append(ids, m.ID)
				}
				_, err := bot.client.ViewMessages(c.ID, 0, ids, true)
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

*/

func (bot *Bot) GetMessage(chatID int64, messageID int64) (*tdlib.Message, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	msg, err := bot.client.GetMessage(chatID, messageID)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return msg, nil
}

// Получить ссылку на сообщение
// @chatID - id чата сообщения
// @messageID - id сообщения
func (bot *Bot) GetMessageLink(chatID int64, messageID int64) (*tdlib.MessageLink, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	ml, err := bot.client.GetMessageLink(chatID, messageID, 0, true, true)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return ml, nil
}

// Получить последнее сообщение в чате
// @chat - целевой чат
// @limit - максимальное количесто сообщений
func (bot *Bot) GetLastMessages(chat *tdlib.Chat, limit int32) ([]tdlib.Message, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if chat == nil || chat.LastMessage == nil {
		return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "No last message ID")
	}

	msgs, err := bot.client.GetChatHistory(chat.ID, chat.LastMessage.ID, -1, limit, false)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	if len(msgs.Messages) == 0 {
		return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", "No message")
	}

	var result []tdlib.Message
	var ids []int64
	for _, m := range msgs.Messages {
		senderID := getSenderID(m.Sender)
		if senderID == bot.Profile.User.ID {
			continue
		}
		//bot.GetMessage(chat.ID, m.ID)
		//bot.client.OpenMessageContent(chat.ID, m.ID)

		ids = append(ids, m.ID)
		result = append(result, m)

	}

	_, err = bot.client.ViewMessages(chat.ID, 0, ids, true)
	if err != nil {
		bot.Logger.Errorln(err)
	}

	bot.GetChatFullInfo(chat.ID)
	//bot.client.GetChat(chat.ID)

	return result, nil

	/*
		msg := msgs.Messages[0]
		// Помечаем сообщения как прочитаные
		// TODO : не работает
		_, err = bot.client.ViewMessages(chat.ID, 0, []int64{msg.ID}, false)
		if err != nil {
			bot.Logger.Errorln(err)
		}

		bot.client.GetChat(chat.ID)

		return &msg, nil
	*/
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
		ft, _ := bot.client.ParseMarkdown(tdlib.NewFormattedText(msg, nil))
		inputMsgTxt := tdlib.NewInputMessageText(ft, true, false)
		mm, e := bot.client.SendMessage(chat.ID, 0, 0, nil, nil, inputMsgTxt)
		fmt.Printf("MES %#v\n\n", mm)
		if e != nil {
			return e.(*tdlib.Error)
		}
	*/

	mid, err := bot.SendMessageByCID(chat.ID, msg)
	if leave {
		bot.client.LeaveChat(chat.ID)
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
	ft, _ := bot.client.ParseMarkdown(tdlib.NewFormattedText(msg, nil))
	inputMsgTxt := tdlib.NewInputMessageText(ft, true, false)

	m, err := bot.client.SendMessage(cid, 0, 0, nil, nil, inputMsgTxt)

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
	chat, err := bot.client.CreatePrivateChat(uid, true)
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
	chat, err := bot.client.CreateNewSupergroupChat(name, true, description, nil, false)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	if address != "" {
		_, err = bot.client.SetSupergroupUsername(chat.Type.(*tdlib.ChatTypeSupergroup).SupergroupID, address)
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
	chats, err := bot.client.SearchPublicChats(query)

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

// Получить полную информацию о чате
// @cid - чат id
func (bot *Bot) GetChatFullInfo(cid int64) (*chat.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chatInfo, err := bot.client.GetChat(cid)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	var result *chat.Chat
	chatType := GetChatType(chatInfo)
	switch chatInfo.Type.GetChatTypeEnum() {
	case tdlib.ChatTypeBasicGroupType:
		group, err := bot.client.GetBasicGroup(chatInfo.Type.(*tdlib.ChatTypeBasicGroup).BasicGroupID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		result = chat.New(chatInfo.ID, chatInfo.Title, "", chatType)
		result.DateCreation = 0
		result.HasLinkedChat = false
		result.IsScam = false
		result.IsVerified = false
		result.MemberCount = group.MemberCount

		//Не корректно отображает дату последнего сообщения
		if chatInfo.LastMessage != nil {
			result.DateLastMessage = chatInfo.LastMessage.Date
		}

		// get caht description
		f, err := bot.client.GetBasicGroupFullInfo(group.ID)
		if err == nil {
			result.BIO = f.Description
		}

	case tdlib.ChatTypeSupergroupType:
		//supergroup
		superGroup, err := bot.client.GetSupergroup(chatInfo.Type.(*tdlib.ChatTypeSupergroup).SupergroupID)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		result = chat.New(chatInfo.ID, chatInfo.Title, superGroup.Usernames.EditableUsername, chatType)
		result.DateCreation = superGroup.Date
		result.HasLinkedChat = superGroup.HasLinkedChat
		result.IsScam = superGroup.IsScam
		result.IsVerified = superGroup.IsVerified
		result.MemberCount = superGroup.MemberCount

		//Не корректно отображает дату последнего сообщения
		if chatInfo.LastMessage != nil {
			result.DateLastMessage = chatInfo.LastMessage.Date
		}

		// get caht description
		f, err := bot.client.GetSupergroupFullInfo(superGroup.ID)
		if err == nil {
			result.BIO = f.Description
		}

	case tdlib.ChatTypePrivateType:
		//user action...
		user, _ := bot.GetUser(chatInfo.ID)
		name := user.FirstName + " " + user.LastName
		username := strings.ToLower(user.Usernames.EditableUsername)
		// Проверка, является ли пользователь ботом
		if strings.HasSuffix(username, "bot") {
			chatType = chat.TypeBot
		}
		result = chat.New(chatInfo.ID, name, username, chatType)

		//fmt.Printf("TYPE : %s\nINFO : %#v\n", chatType, chatInfo)
		//return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Dont supported chat type ChatTypePrivateType")

	default:
		bot.Logger.Infof("UNKNOWN CHAT TYPE:\n\n%#v\n\n", chatInfo)
	}
	if len(chatInfo.Positions) != 0 {
		result.Joined = true
	}
	result.Address = strings.ToLower(result.Address)
	return result, nil
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
	userChat, err := bot.client.SearchPublicChat(username)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
		return 0, e
	}
	_, err = bot.client.AddChatMember(destChat.ID, userChat.ID, 100)
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
	userChat, err := bot.client.SearchPublicChat(username)
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
	_, err = bot.client.SetChatMemberStatus(destChat.ID, member, status)
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

	users, err := bot.client.SearchContacts(username, 10)
	if err != nil {
		e := err.(*tdlib.Error)
		bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
		return 0, e
	}

	var userID int64

	if users.TotalCount == 0 {
		userChat, err := bot.client.SearchPublicChat(username)
		if err != nil {
			e := err.(*tdlib.Error)
			bot.Logger.Debugf("Invite user %s - %s", username, e.Message)
			return 0, e
		}
		_, e := bot.AddContactByID(userChat.ID, username, "")
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

	_, err = bot.client.AddChatMember(destChat.ID, userID, 100)
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
	members, err := bot.client.SearchChatMembers(sourceChat.ID, name, 200, filter)
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
		_, err = bot.client.SearchContacts(name, 10)
		if err != nil {
			e := err.(*tdlib.Error)
			bot.Logger.Debugf("Invite user %s - %s", name, e.Message)
			return 0, e
		}
	*/
	_, e = bot.AddContactByID(uid, name, "")
	if e != nil {
		return 0, e
	}

	destChat, e := bot.GetChat(chatNameDest, true)
	if e != nil {
		return 0, e
	}

	_, err = bot.client.AddChatMember(destChat.ID, uid, 100)
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

	_, err := bot.client.AddChatMember(destChat.ID, uid, 100)
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
		userChat, err := bot.client.SearchPublicChat(userchat)
		if err != nil {
			return err.(*tdlib.Error)
		}
	*/

	_, err := bot.client.AddChatMember(cid, uid, 100)
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
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	time.Sleep(time.Millisecond * time.Duration(timeout))
	member := tdlib.NewMessageSenderUser(uid)
	m, err := bot.client.GetChatMember(cid, member)
	if err != nil {
		//bot.Logger.Errorf("%#v\n", err)
		return err.(*tdlib.Error)
	}
	if m.Status.GetChatMemberStatusEnum() == tdlib.ChatMemberStatusLeftType {
		// вводим новое событие
		ev := tdc.NewEvent(tdc.EventTypeResponse, EventNameChatMemberLeft, 0, "")
		bot.client.PublishEvent(ev)
		return tdlib.NewError(ErrorCodeChatMemberLeft, "CHAT_MEMBER_LEFT", "Telegram auto remove member from chat")
	}

	return nil
}

// Получить инфу о пользователе
// @userID - id пользователя
func (bot *Bot) GetUser(userID int64) (*tdlib.User, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	u, err := bot.client.GetUser(userID)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return u, nil
}

// Получить полную информацию о пользователе. Переделать что бы брать инфу по username!!!
// @uid - ID пользователя
// @chatName - имя чата в которой находится пользователь
func (bot *Bot) GetUserByName(userName string) (*tdlib.User, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}

	chat, errTd := bot.GetChat(userName, false)
	if errTd != nil {
		return nil, errTd
	}

	return bot.GetUser(chat.ID)
}

// Получить полную информацию о пользователе по его номеру телефона.
// @uid - ID пользователя
// @chatName - имя чата в которой находится пользователь
func (bot *Bot) GetUserByPhone(phone string) (*tdlib.User, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}

	firstName := profile.RandomString(5)
	lastName := profile.RandomString(7)
	//проверить существование пользователя
	userID, importErr := bot.ImportContact(phone, firstName, lastName)
	// ErrorUserExists пользователь есть в контактах
	if importErr != nil && importErr.Code != ErrorCodeContactDuplicate {
		return nil, importErr
	}
	bot.client.RemoveContacts([]int64{userID})
	u, err := bot.client.GetUser(userID)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return u, nil
}

// Получить полную информацию о пользователях по списку телефонов
// @uid - ID пользователя
// @chatName - имя чата в которой находится пользователь
func (bot *Bot) GetUserByPhones(phones []string) ([]*tdlib.User, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	//проверить существование пользователя
	userIDs, importErr := bot.ImportContacts(phones)
	// ErrorUserExists пользователь есть в контактах
	if len(userIDs) == 0 {
		return nil, tdlib.NewError(ErrorCodeUserNotExists, "BOT_USER_NOT_EXISTS", "Users not exists")
	}
	if importErr != nil && importErr.Code != ErrorCodeContactDuplicate {
		return nil, importErr
	}

	var result []*tdlib.User

	for _, userID := range userIDs {
		u, err := bot.client.GetUser(userID)
		if err == nil {
			result = append(result, u)
			time.Sleep(time.Second * 1)
		}
	}
	return result, nil
}

// Отправить сообщение участникам группы
func (bot *Bot) SendMessageChatMembers(chatAddr string, message string, offset int32, limit int32) ([]tdlib.ChatMember, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
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
			user, e := bot.client.GetUser(uid)
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
			s, err := bot.client.GetSupergroup(c.SupergroupID)
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
			m, err = bot.client.GetSupergroupMembers(c.SupergroupID, filter, offset, count)
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
		m, err := bot.client.SearchChatMembers(chat.ID, "", 200000, nil)
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

			user, err := bot.client.GetUser(uid)
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
			_, err = bot.client.AddChatMember(destChat.ID, uid, 100)
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

func (bot *Bot) SearchContacts(query string, limit int32) (*tdlib.Users, error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	contacts, err := bot.client.SearchContacts(query, limit)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	return contacts, nil
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

	result, err := bot.client.ImportContacts([]tdlib.Contact{contact})

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

// ImortContacts импортировать список телефонов
// @phones Номера телефонов пользователей
func (bot *Bot) ImportContacts(phones []string) ([]int64, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if len(phones) == 0 {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_WRONG_DATA", "Empty phones list")
	}

	var contacts []tdlib.Contact

	for _, phone := range phones {
		phone = strings.Trim(phone, "\n\t\r ,-:")
		if phone == "" {
			continue
		}
		contact := tdlib.Contact{}
		contact.FirstName = phone
		contact.LastName = ""
		contact.PhoneNumber = phone
		contacts = append(contacts, contact)
	}

	result, err := bot.client.ImportContacts(contacts)
	if err != nil {
		return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", err.Error())
	}

	//bot.Logger.Debugf("Contact %s import success. User ID : %d\n", phone, result.UserIDs[0])
	return result.UserIDs, nil
}

// AddContact добавить контакт по ID. РАБОТАЕТ если пользователь предварительно был найден
// @uid id пользователя
// @firstname Имя пользователя
// @lastname Фамилия пользователя
func (bot *Bot) AddContactByID(uid int64, firstname, lastname string) (int64, *tdlib.Error) {
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

	_, err := bot.client.AddContact(&contact, false)
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

// AddContact добавить в контакты по номеру телефона
// @uid id пользователя
// @firstname Имя пользователя
// @lastname Фамилия пользователя
func (bot *Bot) AddContactByPhone(phone string, firstname, lastname string) (int64, *tdlib.Error) {
	if !bot.IsRun() {
		return 0, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	if phone == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "Empty phone number", "")
	}

	if firstname == "" {
		return 0, tdlib.NewError(ErrorCodeWrongData, "Empty first name", "")
	}

	contact := tdlib.Contact{}
	contact.FirstName = firstname
	contact.LastName = lastname
	contact.PhoneNumber = phone

	_, err := bot.client.AddContact(&contact, true)
	if err != nil {
		return 0, err.(*tdlib.Error)
	}

	if err != nil {
		bot.Logger.Error(err)
		return 0, tdlib.NewError(ErrorCodeUserNotExists, "Error", err.Error())
	}

	bot.Logger.Debugf("Пользователь %s успешно добавлен в контакты", phone)

	//Сделать получение ID
	return 0, nil
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
			bot.client.PublishEvent(ev)
		} else {
			bot.Logger.Error("Import error : ", importErr)
		}
		return 0, importErr
	}
	mid, sendErr := bot.SendMessageByUID(uid, msg, ontime)
	if sendErr == nil {
		bot.Logger.Debugf("Send to %s success. Message ID : %d\n", phone, mid)
	}
	bot.client.PublishEvent(ev)
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

	chats, err := bot.client.SearchPublicChats(username)
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

	usr, err := bot.client.GetUser(uid)
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
	return bot.SendMessageByCID(chat.ID, msg)

}

// Дождаться статус отправки сообщения
func (bot *Bot) WaitMessageState(msg *tdlib.Message) *tdlib.Error {
	//проверяем статус сообщения. Проверить на большом кол-ве сообщений в чате.
	//Переделать!!!(что бы не загружал список всех сообщений. А еще лучше сделать асинхронное получение статуса)
	for {
		time.Sleep(500 * time.Millisecond)

		mm, err := bot.client.GetChatHistory(msg.ChatID, 0, 0, 1, true)
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
			bot.client.PublishEvent(ev)
			return e
		}
		//иначе сообщение в процессе отправки. Ждем 1 сек

	}

	return nil
}

// Загрузить информацию о чате
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

	if !IsPublicLink(chatname) {
		//bot.Logger.Debugln("Parse private chat ", chatname)
		var chatInfo *tdlib.ChatInviteLinkInfo
		chatInfo, err = bot.client.CheckChatInviteLink(chatname)
		if err != nil {
			return nil, tdlib.NewError(ErrorCodeSystem, "BOT_SYSTEM_ERROR", err.Error())
		}
		//Если уже в группе
		if chatInfo.ChatID != 0 {
			bot.Logger.Debugln("This bot is already in the group", chatname)
			chat, _ = bot.client.GetChat(chatInfo.ChatID)

		} else {
			//Вступаем в группу
			chat, err = bot.client.JoinChatByInviteLink(chatname)
			if err != nil {
				bot.Logger.Errorf("Join to %s error: %s\n", chatname, err)
				return nil, err.(*tdlib.Error)
			}
		}
	} else {

		chatname = ExtrctChatName(chatname)
		//Если chatID == int64 тогда ищем чат по id иначе по имени
		cid, err := strconv.ParseInt(chatname, 10, 64)
		if err != nil {
			// Ищем в своих группах
			// TODO: не всегда находит чат
			//bot.client.SearchChats(profile.RandomString(5), 1)
			chts, err := bot.client.SearchChats(chatname, 1)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
			//fmt.Printf("%#v", chts)
			if chts.TotalCount > 0 {
				bot.Logger.Debugln("This bot is already in the group", chatname)
				chat, _ = bot.client.GetChat(chts.ChatIDs[0])
			} else {
				// Ищем публичную группу
				chat, err = bot.client.SearchPublicChat(chatname)
				if err != nil {
					return nil, err.(*tdlib.Error)
				}
			}
		} else {
			//Получаем группу по ID
			chat, err = bot.client.GetChat(cid)
			if err != nil {
				return nil, err.(*tdlib.Error)
			}
		}
		if join && chat.Type.GetChatTypeEnum() != tdlib.ChatTypePrivateType {
			_, err := bot.client.JoinChat(chat.ID)
			if err != nil {
				bot.Logger.Errorf("Join to %s error: %s\n", chat.ID, err)
				return nil, err.(*tdlib.Error)
			}
		}
	}
	//time.Sleep(time.Millisecond * 500)
	return chat, nil

}

func (bot *Bot) GetChatByID(chatID int64) (*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chat, err := bot.client.GetChat(chatID)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	return chat, nil
}

// Список всех созданных пользователем чатов.
func (bot *Bot) GetCreatedChats() ([]*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chatType := tdlib.NewPublicChatTypeHasUsername()
	chats, err := bot.client.GetCreatedPublicChats(chatType)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	var result []*tdlib.Chat
	for _, cid := range chats.ChatIDs {
		chat, err := bot.client.GetChat(cid)
		if err != nil {
			return nil, err.(*tdlib.Error)
		}
		result = append(result, chat)
	}
	return result, nil

}

func (bot *Bot) DeleteChat(chatID int64) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	_, err := bot.client.DeleteChat(chatID)
	if err != nil {
		return err.(*tdlib.Error)
	}
	return nil
}

func (bot *Bot) LeaveChat(chatID int64) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	_, err := bot.client.LeaveChat(chatID)
	if err != nil {
		return err.(*tdlib.Error)
	}
	return nil
}

// Список всех созданных пользователем каналов.
// TODO: отображаются только публичные. Сделать что бы подгружались и скрытые
func (bot *Bot) GetCreatedChannels() ([]*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chatType := tdlib.NewPublicChatTypeHasUsername()
	chats, err := bot.client.GetCreatedPublicChats(chatType)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	var result []*tdlib.Chat
	for _, cid := range chats.ChatIDs {
		chat, err := bot.client.GetChat(cid)
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
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chatType := tdlib.NewPublicChatTypeHasUsername()
	chats, err := bot.client.GetCreatedPublicChats(chatType)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}
	var result []*tdlib.Chat
	for _, cid := range chats.ChatIDs {
		chat, err := bot.client.GetChat(cid)
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

func (bot *Bot) SearchPublicChat(username string) (*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chat, err := bot.client.SearchPublicChat(username)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	return chat, nil
}

// Получаем список чатов аккаунта. Результат записываем в bot.Chats и возвращаем в ответ
// TODO: проверить загружает ли все чаты. Обдумать нужно ли возвращать список
func (bot *Bot) GetChatList(limit int32) ([]*tdlib.Chat, *tdlib.Error) {
	if !bot.IsRun() {
		return nil, tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	chats, err := bot.client.GetChats(nil, limit)
	if err != nil {
		return nil, err.(*tdlib.Error)
	}

	var result []*tdlib.Chat
	for _, chatID := range chats.ChatIDs {
		// get chat info from tdlib
		chat, err := bot.client.GetChat(chatID)
		if err == nil {
			result = append(result, chat)
		} else {
			return nil, err.(*tdlib.Error)
		}
	}
	return result, nil
}

// Жалоба на чат
func (bot *Bot) ReportChat(chatID int64, messageIDs []int64, reason tdlib.ChatReportReason, text string) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	_, e := bot.client.ReportChat(chatID, messageIDs, reason, text)
	if e != nil {
		return e.(*tdlib.Error)
	}
	return nil
}

func (bot *Bot) LeaveChatList(chats []*tdlib.Chat, limit int32) *tdlib.Error {
	if !bot.IsRun() {
		return tdlib.NewError(ErrorCodeWrongData, "BOT_SYSTEM_ERROR", "Bot dying")
	}
	var current int32
	var e error
	for _, c := range chats {
		if current >= limit {
			break
		}
		//fmt.Printf("Leave Chat error %#v\n", c.Permissions.CanChangeInfo)
		if !c.Permissions.CanChangeInfo {
			switch c.Type.GetChatTypeEnum() {
			case tdlib.ChatTypePrivateType:
				_, e = bot.client.DeleteChat(c.ID)
				//continue
			case tdlib.ChatTypeBasicGroupType:
				_, e = bot.client.DeleteChat(c.ID)
			case tdlib.ChatTypeSupergroupType:
				_, e = bot.client.LeaveChat(c.ID)
			default:
				e = fmt.Errorf("Unknown chat type")
			}
			if e != nil {
				fmt.Println("Leave Chat error ", e)
			} else {
				current++
			}
		}

	}

	return nil
}
