package client

import "time"

func run() {
	c := NewClient(1)
	c.Connect()
	c.List()
	c.Stor(1024 * 1024 * 200)
	c.Stor(1024 * 1024 * 200)
	c.Stor(1024 * 1024 * 200)
	c.Stor(1024 * 1024 * 200)
}

func StressTest() {
	for {
		time.Sleep(1 * time.Second)
		go run()
		c := NewClient(1)
		c.Connect()
		c.List()
		c.Stor(1024 * 1024 * 200)
		c.List()
		c.Stor(1024 * 1024 * 400)
		c.List()
		go run()
	}
}
