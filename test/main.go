package main

import (
	"fmt"

	"github.com/juju/fslock"
	"github.com/polihoster/tdbot"
	"github.com/polihoster/tdbot/config"
	"github.com/polihoster/tdbot/profile"
)

func main() {

	/*	conf, err := config.Load("bot.toml")
		if err != nil {
			fmt.Println(err)
			return
		}
	*/

	prof, err := profile.Get("79111024853", 0)
	//	prof, err := profile.New("79508636519", "", conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot := tdbot.New(prof)

	go func() {
		err = bot.Start()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	//	time.Sleep(time.Second * 10)
	//	bot.Stop()
	var code string
	fmt.Scanf("%s\n", &code)

	evns, err := bot.Profile.Event.Search("", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, ev := range evns {
		fmt.Printf("%s : %s ::::: %s\n", ev.Type, ev.Name, ev.Data)
	}
	//bot.SendCode(code)
	fmt.Scanf("%s", &code)
}

func testMove() {

	conf := config.New()

	err := conf.Load("bot.toml")
	if err != nil {
		fmt.Println(err)
		return
	}

	prof, err := profile.New("123456", "./", conf)
	//prof, err := profile.Get("./123456")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = prof.Move("tmp")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n", prof)
}

/*
func testLock() {

	//runtime.BlockProfile()

	go func() {
		for {

			err := profile.LockProfile("prof")
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Lock OK")
			}
			time.Sleep(time.Second * 1)

		}
	}()

	go func() {
		for {
			time.Sleep(time.Second * 10)
			err := profile.UnlockProfile("prof")
			if err != nil {
				fmt.Println("UNLOCK : ", err)
			} else {
				fmt.Println("Unlock OK")
			}
		}
	}()

	var name string
	fmt.Println("Run...")
	fmt.Scanf("%s\n", &name)

}
*/

func lock() {
	lock := fslock.New("lock")
	err := lock.TryLock()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("OK")
}
