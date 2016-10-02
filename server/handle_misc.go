package server

func (c *clientHandler) handleSYST() {
	c.writeMessage(215, "UNIX Type: L8")
}

func (c *clientHandler) handleFEAT() {
	c.writer.WriteString("211- These are my features\r\n")

	c.writer.WriteString(" UTF8\r\n")

	c.writeMessage(211, "end")
}

func (c *clientHandler) handleTYPE() {
	c.writeMessage(200, "Type set to binary")
}

func (c *clientHandler) handleQUIT() {
	//fmt.Println("Goodbye")
	c.writeMessage(221, "Goodbye")
	c.disconnect()
}

func (c *clientHandler) handleSTAT() {
	c.writeMessage(551, "downloads not allowed")
}
