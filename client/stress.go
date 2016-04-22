package client

import "time"

func StressTest() {
	time.Sleep(1 * time.Second)
	c := NewClient(1)
	c.Connect()
	c.List()
	c.Stor(1024)
	time.Sleep(1 * time.Second)
	c.List()
	time.Sleep(1 * time.Second)
	c.Stor(1024 * 20)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	c.List()
}
