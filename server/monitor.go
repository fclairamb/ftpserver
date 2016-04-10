package server

import "fmt"
import "time"

func Monitor() {
	for {
		fmt.Println(" * MONITOR *")
		time.Sleep(1 * time.Second)
	}
}
