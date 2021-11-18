// Package user внутреннее представление пользователя
package user

//User структура данных пользователя
type User struct {
	ID          int32  `json:"id"`
	PhoneNumber string `json:"phone"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Addr        string `json:"addr"` //для телеграм @username
	WasOnline   int32  `json:"was_online"`
	Status      Status `json:"status"` //статус в сети телеграм
	//Health      Health `json:"health"` //текущее состояние
	Location string `json:"location"`
	Type     Type   `json:"type"`
}

// Type ...
type Type string

// Status ...
type Status string

// Health
type Health string

var (
	//TypeTelegram user
	TypeTelegram Type = "telegram"
	//TypeFacebook user
	TypeFacebook Type = "facebook"
)

var (
	StatusInitialization Status = "init"
	//StatusSendPhone      Status = "send_phone"
	//StatusWaitPhone      Status = "wait_phone"
	//StatusWaitCode       Status = "wait_code"
	//StatusWaitPassword   Status = "wait_pass"
	//StatusRegistration   Status = "registration"
	//StatusRestricted у пользователя есть ограничения
	StatusRestricted Status = "restricted"
	// StatusReady ....
	StatusReady Status = "ready"
	// StatusOffline ....
	StatusOffline Status = "offline"
	// StatusUnknown ....
	StatusUnknown Status = "unknown"
	// StatusNotExist ....
	StatusNotExist Status = "notexist"
	// StatusBanned ....
	StatusBanned Status = "banned"
	// StatusLogout ....
	StatusLogout Status = "logout"

	StatusStopped Status = "stopped"
	// Limit exceeded
	//StatusLimitExceeded Status = "limit"
)

var (
	HealthGood  Health = "good"
	HealthBan   Health = "ban"
	HealthLimit Health = "limit"
)

// New ...
func New(name, phone string, usrType Type) *User {

	return &User{
		ID:          0,
		FirstName:   name,
		LastName:    "",
		Addr:        "",
		PhoneNumber: phone,
		Status:      StatusUnknown,
		Type:        usrType,
		//WasOnline:   ???? //доделать получение времени активности
	}

}
