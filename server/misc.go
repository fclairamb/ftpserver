package server

import (
	//"fmt"
	"strings"
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

func (self *Paradise) handleCwd() {
	if self.param == ".." {
		self.path = "/"
	} else {
		self.path = self.param
	}
	if !strings.HasPrefix(self.path, "/") {
		self.path = "/" + self.path
	}
	self.writeMessage(250, "CD worked")
}

func (self *Paradise) handleSize() {
	self.writeMessage(450, "downloads not allowed")
}

func (self *Paradise) handleRetr() {
	self.writeMessage(551, "downloads not allowed")
}

func (self *Paradise) handleStat() {
	self.writeMessage(551, "downloads not allowed")
}
