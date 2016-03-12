package server

import (
//"fmt"
)

func (self *Paradise) handleSyst() {
	self.writeMessage(215, "UNIX Type: L8")
}

func (self *Paradise) handlePwd() {
	self.writeMessage(257, "\""+self.path+"\" is the current directory")
}

func (self *Paradise) handleType() {
	self.writeMessage(200, "Type set to binary")
}

func (self *Paradise) handleQuit() {
	self.writeMessage(221, "Goodbye")
	self.theConnection.Close()
}
