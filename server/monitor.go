package server

import "fmt"
import "time"

func Monitor() {
	for {
		fmt.Println("9 clients, 4 passive, Up for 12:33:12.34")
		fmt.Println("   PFC106 00:17:34")
		fmt.Println("   PFC101 00:09:04")
		fmt.Println("   PFC109 00:07:04")
		fmt.Println("     PFC109p 03:33")
		fmt.Println("     PFC109p 01:33")
		fmt.Println("     PFC109p 00:33")
		fmt.Println("   PFC116 00:06:04")
		fmt.Println("   PFC126 00:05:30")
		fmt.Println("   PFC104 00:04:30")
		fmt.Println("   PFC206 00:03:30")
		fmt.Println("     PFC206p 00:02:30")
		fmt.Println("   PFC306 00:02:30")
		fmt.Println("   PFC301 00:01:30")
		time.Sleep(5 * time.Second)
	}
}
