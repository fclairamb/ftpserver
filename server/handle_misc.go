package server

func (c *ClientHandler) handleSYST() {
	c.writeMessage(215, "UNIX Type: L8")
}

func (c *ClientHandler) handleTYPE() {
	c.writeMessage(200, "Type set to binary")
}

func (c *ClientHandler) handleQUIT() {
	//fmt.Println("Goodbye")
	c.writeMessage(221, "Goodbye")
	c.disconnect()
}


func (c *ClientHandler) handleSTAT() {
	c.writeMessage(551, "downloads not allowed")
}
