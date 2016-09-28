package server

func (c *ClientHandler) HandleSYST() {
	c.writeMessage(215, "UNIX Type: L8")
}

func (c *ClientHandler) HandleTYPE() {
	c.writeMessage(200, "Type set to binary")
}

func (c *ClientHandler) HandleQUIT() {
	//fmt.Println("Goodbye")
	c.writeMessage(221, "Goodbye")
	c.Die()
}


func (c *ClientHandler) HandleSTAT() {
	c.writeMessage(551, "downloads not allowed")
}
