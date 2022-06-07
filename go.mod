module github.com/satyshef/tdbot

go 1.15

// For local develop enable replace

//replace github.com/satyshef/tdlib => ../tdlib

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/satyshef/tdlib v0.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
)
