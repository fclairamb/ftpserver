package client

import "time"

func StressTest() {
	time.Sleep(3 * time.Second)
	c := NewClient(1)
	c.Connect()
	c.List()
	c.Stor(1024)
}
