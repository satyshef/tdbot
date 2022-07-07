package profile

import (
	"github.com/satyshef/tdbot/config"
	"github.com/satyshef/tdbot/events/eventman"
	"github.com/satyshef/tdbot/user"
)

//Profile ...
type Profile struct {
	User   *user.User
	dir    string
	Config *config.Config
	Event  *eventman.Manager
}

type Sort int

const (
	SORT_ALPHABET  Sort = 0
	SORT_RANDOM    Sort = 1
	SORT_TIME_ASC  Sort = 2
	SORT_TIME_DESC Sort = 3
)

/*
const (
	EventTypeRequest  = "request"
	EventTypeResponse = "response"
	EventTypeError    = "error"
)
*/
const (
	//коды ошибок данного модуля начинаются с 2
	ErrorCodeLimitExceeded = 201
	ErrorCodeLimitNotSet   = 202
	ErrorCodeNotInit       = 203
	ErrorCodeDirNotExists  = 204
	ErrorCodeSystem        = 205
)
