package main

import (
	"fmt"
	"time"

	"github.com/juju/fslock"
	"github.com/polihoster/tdbot/profile"
)

func main() {

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

func lock() {
	lock := fslock.New("lock")
	err := lock.TryLock()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("OK")
}
