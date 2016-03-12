package server

import (
	"fmt"
)

func (self *Paradise) handleUser() {
	fmt.Println(self.ip, self.command, self.param)
	self.writeMessage(331, "User name ok, password required")
}

func (self *Paradise) handlePass() {
}
