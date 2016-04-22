package server

func (p *Paradise) HandleUser() {
	p.user = p.param
	p.writeMessage(331, "User name ok, password required")
}

func (p *Paradise) HandlePass() {
	// think about using https://developer.bitium.com
	if AuthManager.CheckUser(p.user, p.param) {
		p.writeMessage(230, "Password ok, continue")
	} else {
		p.writeMessage(530, "Incorrect password, not logged in")
	}
}
