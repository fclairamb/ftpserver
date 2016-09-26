package server

func (c *ClientHandler) HandleRetr() {
	c.writeMessage(551, "downloads not allowed")
}
