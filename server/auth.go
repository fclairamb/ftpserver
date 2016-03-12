package server

import (
	"fmt"
)

func (self *Paradise) handleUser() {
	fmt.Println(self.ip, self.command, self.param)
	self.user = self.param
	self.writeMessage(331, "User name ok, password required")
}

func (self *Paradise) handlePass() {
	if true { // change to your auth logic, think about using https://developer.bitium.com
		self.writeMessage(230, "Password ok, continue")
	} else {
		self.writeMessage(530, "Incorrect password, not logged in")
	}
}
