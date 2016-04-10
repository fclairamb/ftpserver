package main

import "testing"
import "os"

import "os/exec"
import "paradise/client"
import "math/rand"
import "sync"
import "time"

var file *os.File
var fileBytes []byte

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestSimple(t *testing.T) {
	cmd := exec.Command("killall", "paradise")
	cmd.Run()
	cmd = exec.Command("./paradise")
	cmd.Start()
	time.Sleep(3 * (time.Second * 1))
	testConnect(t)
	testLots(t)
}

func testConnect(t *testing.T) {
	c := client.NewClient(1)
	c.Connect()
	c.List()
	c.Stor(1024)
}

func testLots(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	for i := 0; i < 5; i++ {
		go randClient(i)
		time.Sleep((time.Millisecond * 500))
	}
	wg.Wait()
}

func randClient(id int) {
	c := client.NewClient(id)
	c.Connect()
	for {
		c.List()
		c.Stor(int64(1024 * 1024 * rand.Intn(20)))
		time.Sleep((time.Millisecond * 500))
	}
}
