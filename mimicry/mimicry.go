package mimicry

import (
	"fmt"

	"github.com/satyshef/tdbot"
	"github.com/satyshef/tdbot/user"
	"github.com/satyshef/tdlib"

	"github.com/sirupsen/logrus"
)

// Human ...
type Human struct {
	Bot     *tdbot.Bot
	Friends []*user.User
	Logger  *logrus.Logger
}

// NewHuman ...
func NewHuman(bot *tdbot.Bot) *Human {
	return &Human{
		Bot:     bot,
		Friends: *new([]*user.User),
		Logger:  logrus.New(),
	}
}

// Получить список друзей
func (h *Human) GetFriendsList() ([]int64, error) {
	if h.Bot.Profile.Config.Mimicry == nil || h.Bot.Profile.Config.Mimicry.FriendName == "" || h.Bot.Profile.Config.Mimicry.FriendCount == 0 {
		return nil, fmt.Errorf("Mimicry not set")
	}

	friends, err := h.Bot.Client.SearchContacts(h.Bot.Profile.Config.Mimicry.FriendName, int32(h.Bot.Profile.Config.Mimicry.FriendCount))
	if err != nil {
		return nil, err
	}

	var ids []int64

	//Если друзей меньше указанного количества то найти новых
	if len(friends.UserIDs) < h.Bot.Profile.Config.Mimicry.FriendCount {
		newFiendsIDs, err := h.MakeFriends(h.Bot.Profile.Config.Mimicry.FriendCount - len(friends.UserIDs))
		if err != nil {
			return nil, err
		}
		ids = append(friends.UserIDs, newFiendsIDs...)
	} else {
		ids = friends.UserIDs
	}

	return ids, nil
}

func (h *Human) GetRandomFriend() (int64, error) {
	ids, err := h.GetFriendsList()
	if err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, fmt.Errorf("Empty friend list")
	}

	i := RandInt(0, len(ids))
	return ids[i], nil
}

//Подружиться
func (h *Human) MakeFriends(count int) ([]int64, error) {
	var memberLimit int32 = 200

	//Друзей берем из группы
	//chat, err := h.GetFriendChat()
	chat, err := h.Bot.GetChat(h.Bot.Profile.Config.Mimicry.FriendRoom, true)
	if err != nil {
		return nil, err
	}

	//Пишем приветственное сообщение
	h.Bot.SendMessageByCID(chat.ID, fmt.Sprintf("Hi i'm %s", h.Bot.Profile.User.FirstName))

	var members *tdlib.ChatMembers
	if chat.Type.GetChatTypeEnum() == tdlib.ChatTypeSupergroupType {
		c := chat.Type.(*tdlib.ChatTypeSupergroup)
		members, _ = h.Bot.Client.GetSupergroupMembers(c.SupergroupID, nil, 0, memberLimit)
	} else {
		members, _ = h.Bot.Client.SearchChatMembers(chat.ID, "", memberLimit, nil)
	}
	result := []int64{}

	//добавляем ботов-друзей в контакты и записываем их id
	for i := 0; i < count; i++ {
		//берем случайного бота
		m := RandInt(0, int(members.TotalCount)-1)
		//игнорируем себя
		if h.Bot.Profile.User.ID == members.Members[m].MemberID.GetID() {
			continue
		}
		//игнорируем создателя и администратора
		if members.Members[m].Status.GetChatMemberStatusEnum() == tdlib.ChatMemberStatusCreatorType || members.Members[m].Status.GetChatMemberStatusEnum() == tdlib.ChatMemberStatusAdministratorType || members.Members[m].Status.GetChatMemberStatusEnum() == tdlib.ChatMemberStatusBannedType {
			continue
		}
		//fmt.Printf("MEMBER : %#v\n\n", members.Members[m])
		h.Bot.AddContact(members.Members[m].MemberID.GetID(), h.Bot.Profile.Config.Mimicry.FriendName, fmt.Sprintf("%d", members.Members[m].MemberID.GetID()))
		result = append(result, members.Members[m].MemberID.GetID())
	}

	return result, nil

}

/*
//Подключение к дружественным группам
func (h *Human) GetFriendChat() (*tdlib.Chat, *tdlib.Error) {


	if h.Bot.Profile.Config.Mimicry.FriendRoom == "" {
		return nil, tdlib.NewError(profile.ErrorCodeSystem, "PROFILE_WRONG_PARAM", "Empty friend room link")
	}

	chatInfo, err := h.Bot.Client.CheckChatInviteLink(h.Bot.Profile.Config.Mimicry.FriendRoom)
	if err != nil {
		//h.Bot.Logger.Errorf("Check FriendRoom link error: %s\n", err)
		return nil, err.(*tdlib.Error)
	}

	//fmt.Printf("INFO %#v\n", chatInfo)

	//Если уже в группе
	if chatInfo.ChatID != 0 {
		h.Bot.Logger.Debugln("This bot is already in the friend group")
		chat, _ := h.Bot.Client.GetChat(chatInfo.ChatID)
		//h.Bot.Logger.Debugf("G : %#v\n", chat)
		return chat, nil
	}

	//Вступаем в группу
	chat, err := h.Bot.Client.JoinChatByInviteLink(h.Bot.Profile.Config.Mimicry.FriendRoom)
	if err != nil {
		h.Bot.Logger.Errorf("Join to %s error: %s\n", h.Bot.Profile.Config.Mimicry.FriendRoom, err)
		return nil, tdlib.NewError(tgbot.ErrorCodeSystem, "BOT_SYSTEM_ERROR", fmt.Sprintf("Join to %s error: %s\n", h.Bot.Profile.Config.Mimicry.FriendRoom, err))
	}

	h.Bot.Logger.Debugf("Join to group %s success %#v\n", chat.Title, chat)

	//Пишем приветственное сообщение
	h.Bot.SendMessageByCID(chat.ID, fmt.Sprintf("Hi i'm %s", h.Bot.Profile.User.FirstName))
	return chat, nil

}
*/

/*
// Start инициализация человеческого поведения
// Бот подключается к группам, и ведет переписку с участниками данных групп
func (h *Human) Start() {
	go h.Init()
}

// Init ...
func (h *Human) Init() {

	h.Logger.SetLevel(6)

	//Ждем готовности бота
	for !h.Bot.IsReady() {
		time.Sleep(time.Second * 1)
	}

	go h.start()
}

// инициализация переписки с друзьями
func (h *Human) start() {

	if h.Bot.Profile.Config.Humanity == nil {
		h.Logger.Debugf("У бота %s нет настроек человечности", h.Bot.Profile.User.PhoneNumber)
		return
	}
	for true {
		//Интервал отправки сообщений в минутах
		interval := randInt(h.Bot.Profile.Config.Humanity.Friend.SendMin, h.Bot.Profile.Config.Humanity.Friend.SendMax)
		time.Sleep(time.Minute * time.Duration(interval))

		//Загружаем список друзей
		h.Friends = h.Bot.LoadFriends()
		h.SendFriendMessage("Rebellion is coming")
	}

}

// SendFriendMessage отправить случайному другу сообщение
func (h *Human) SendFriendMessage(msg string) {

	//Нет друзей
	if len(h.Friends) == 0 {
		h.Logger.Debugf("У бота %s нет друзей", h.Bot.Profile.User.PhoneNumber)
		return
	}

	//Отправить случайному другу сообщение
	//Выбираем случайного пользователя из списка друзей
	i := randInt(0, len(h.Friends)-1)
	friend := h.Friends[i]

	h.Bot.SendMessageByUID(friend.ID, msg, 0)
}
*/
/*
//SetFriendRooms установить группы друзей
func (bot *Bot) SetFriendRooms(friendRooms []config.FriendRoom) {
	bot.Profile.Config.Humanity.Friend.Rooms = friendRooms
}

// CreatePrivateChan создать канал для логирования запросов бота
func (bot *Bot) CreatePrivateChan(chanName string) error {
	var err error
	var privateChan *tdlib.Chat

	//t.Client.GetRawUpdatesChannel(1000)

	chats, err := bot.GetChatList(1000)
	if err != nil {
		return err
	}

	for _, ch := range chats {
		if ch.Title == chanName {
			privateChan = ch
			break
		}
	}

	//если канал не существует тогда создаем его
	if privateChan == nil {

		//создаем канал для логирования действий бота
		privateChan, err = bot.Client.CreateNewSupergroupChat(chanName, true, "", nil)
		if err != nil {
			//tdErr := ErrorDecode(err.Error())
			bot.Logger.Errorf("Ошибка при создании личного канала бота : %s", err)
			//os.Exit(1)
			return nil
		}

		if privateChan == nil {
			bot.Logger.Errorf("Ошибка при создании личного канала бота %s : пустое значение logChan", chanName)
			os.Exit(1)
		}

		bot.Logger.Infof("Личный канал \"%s\" бота успешно создан\n", chanName)
	} else {
		bot.Logger.Infof("Канал \"%s\" существует\n", privateChan.Title)
	}

	bot.PrivateChan = privateChan

	return nil

}
*/
