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
