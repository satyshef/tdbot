package mimicry

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

//RandInt случайное число int
func RandInt(min, max int) int {

	if min > max {
		fmt.Printf("Incorrect interval value: min %d - max %d", min, max)
		os.Exit(1)
	}

	n := max - min
	if n == 0 {
		return min
	}

	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(n)
}

func RandString(size int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, size)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
