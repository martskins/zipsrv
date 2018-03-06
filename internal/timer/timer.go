package timer

import (
	"log"
	"time"
)

func Timer(funcName string) func() {
	a := time.Now()

	return func() {
		t := time.Since(a).Nanoseconds()
		log.Printf("%s took %v nanoseconds", funcName, t)
	}
}
