package server

import (
	"fmt"
	"io"
)

func (self *Paradise) handleStore() {
	fmt.Println(self.ip, self.command, self.param)

	self.writeMessage(150, "Data transfer starting")

	_, err := self.storeOrAppend()
	if err == io.EOF {
		self.writeMessage(226, "OK, received some bytes") // TODO send total in message
	} else {
		self.writeMessage(550, "Error with upload: "+err.Error())
	}
}

func (self *Paradise) storeOrAppend() (int64, error) {
	err := self.readFirst512Bytes()
	if err != nil {
		return 0, err
	}

	// TODO run self.buffer thru mime type checker
	// if mime type bad, reject upload

	return 0, nil
}

func (self *Paradise) readFirst512Bytes() error {
	self.buffer = make([]byte, 0)
	var err error
	self.waiter.Wait()
	for {
		temp_buffer := make([]byte, 512)
		n, err := self.passiveConn.Read(temp_buffer)

		if err != nil {
			break
		}
		self.buffer = append(self.buffer, temp_buffer[0:n]...)

		if len(self.buffer) >= 512 {
			break
		}
	}

	if err != nil && err != io.EOF {
		return err
	}

	// you have a buffer filled to 512, or less if file is less than 512
	return nil
}
