module github.com/satyshef/tdbot

go 1.15

replace github.com/satyshef/tdlib => ../tdlib

replace github.com/satyshef/tdbot/chat => ./chat

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/satyshef/tdbot/chat v0.0.0-00010101000000-000000000000
	github.com/satyshef/tdlib v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
)
