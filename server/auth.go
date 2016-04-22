package server

func (p *Paradise) HandleUser() {
	//fmt.Println(p.ip, p.command, p.param)
	p.user = p.param
	p.writeMessage(331, "User name ok, password required")
}

func (p *Paradise) HandlePass() {
	if true { // change to your auth logic, think about using https://developer.bitium.com
		p.writeMessage(230, "Password ok, continue")
	} else {
		p.writeMessage(530, "Incorrect password, not logged in")
	}
}
