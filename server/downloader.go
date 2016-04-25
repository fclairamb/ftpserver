package server

func (p *Paradise) HandleRetr() {
	p.writeMessage(551, "downloads not allowed")
}
