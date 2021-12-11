module github.com/polihoster/tdbot

go 1.15

replace github.com/polihoster/tdlib => ../tdlib

replace github.com/polihoster/tdbot/chat => ./chat

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/polihoster/tdbot/chat v0.0.0-00010101000000-000000000000
	github.com/polihoster/tdlib v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
)
