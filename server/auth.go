package server

import "fmt"

func (p *ClientHandler) HandleUser() {
	p.user = p.param
	p.writeMessage(331, "OK")
}

func (p *ClientHandler) HandlePass() {
	// think about using https://developer.bitium.com
	if err := p.daddy.driver.CheckUser(p, p.user, p.param); err == nil {
		p.writeMessage(230, "Password ok, continue")
	} else {
		p.writeMessage(530, fmt.Sprintf("Authentication problem: %s", err))
		p.Die()
	}
}
