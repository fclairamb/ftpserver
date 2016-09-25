package server

func (p *ClientHandler) HandleRetr() {
	p.writeMessage(551, "downloads not allowed")
}
