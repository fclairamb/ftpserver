package server

func (p *ClientHandler) HandleSyst() {
	p.writeMessage(215, "UNIX Type: L8")
}

func (p *ClientHandler) HandleType() {
	p.writeMessage(200, "Type set to binary")
}

func (p *ClientHandler) HandleQuit() {
	//fmt.Println("Goodbye")
	p.writeMessage(221, "Goodbye")
	p.conn.Close()
	delete(p.daddy.ConnectionMap, p.cid)
}

func (p *ClientHandler) HandleSize() {
	p.writeMessage(450, "downloads not allowed")
}

func (p *ClientHandler) HandleStat() {
	p.writeMessage(551, "downloads not allowed")
}
