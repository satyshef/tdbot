module github.com/satyshef/tdbot

go 1.18

// For local develop enable replace

//replace github.com/satyshef/go-tdlib => ../go-tdlib

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/satyshef/go-tdlib v0.3.11
	github.com/sirupsen/logrus v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c

)

require (
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.1.0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
)
