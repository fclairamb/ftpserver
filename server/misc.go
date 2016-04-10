package server

import "fmt"
import "strings"

func (self *Paradise) HandleSyst() {
	self.writeMessage(215, "UNIX Type: L8")
}

func (self *Paradise) HandlePwd() {
	self.writeMessage(257, "\""+self.path+"\" is the current directory")
}

func (self *Paradise) HandleType() {
	self.writeMessage(200, "Type set to binary")
}

func (self *Paradise) HandleQuit() {
	fmt.Println("Goodbye")
	self.writeMessage(221, "Goodbye")
	self.theConnection.Close()
}

func (self *Paradise) HandleCwd() {
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

func (self *Paradise) HandleSize() {
	self.writeMessage(450, "downloads not allowed")
}

func (self *Paradise) HandleRetr() {
	self.writeMessage(551, "downloads not allowed")
}

func (self *Paradise) HandleStat() {
	self.writeMessage(551, "downloads not allowed")
}
