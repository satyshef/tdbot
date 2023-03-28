package chat

import "encoding/json"

type Chat struct {
	/*
		ID              int64    `json:"id"`
		Name            string   `json:"name"`
		Address         string   `json:"addr"`
		Type            Type     `json:"type"`
		MemberCount     int32    `json:"member_count"`
		DateCreation    int32    `json:"date_creation"`
		DateLastMessage int32    `json:"date_last_msg"`
		IsVerified      bool     `json:"is_verified"`
		IsScam          bool     `json:"is_scam"`
		HasLinkedChat   bool     `json:"has_linked_chat"`
		Admins          []string `json:"admins"`
		BIO             string   `json:"bio"`
		Lang            string   `json:"lang_code"`
	*/

	ID              int64
	Name            string
	Address         string
	Type            Type
	MemberCount     int32
	DateCreation    int32
	DateLastMessage int32
	IsVerified      bool
	IsScam          bool
	HasLinkedChat   bool
	Admins          []string
	BIO             string
	Lang            string
}

// Тип чата
type Type string

const (
	TypeChannel        Type = "channel"
	TypePrivateChannel Type = "private_channel"
	TypeGroup          Type = "group"
	TypePrivateGroup   Type = "private_group"
	TypeUser           Type = "user"
	TypeBot            Type = "bot"
)

func New(id int64, name string, addr string, chatType Type) *Chat {
	return &Chat{
		ID:      id,
		Name:    name,
		Address: addr,
		Type:    chatType,
	}
}

func Decode(data []byte) (result *Chat, err error) {

	err = json.Unmarshal(data, result)

	return
}
