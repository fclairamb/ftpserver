package server

func (c *ClientHandler) HandleSyst() {
	c.writeMessage(215, "UNIX Type: L8")
}

func (c *ClientHandler) HandleType() {
	c.writeMessage(200, "Type set to binary")
}

func (c *ClientHandler) HandleQuit() {
	//fmt.Println("Goodbye")
	c.writeMessage(221, "Goodbye")
	c.Die()
}


func (c *ClientHandler) HandleStat() {
	c.writeMessage(551, "downloads not allowed")
}
