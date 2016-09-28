package server

import "fmt"

// Handle the "USER" command
func (p *ClientHandler) HandleUSER() {
	p.user = p.param
	p.writeMessage(331, "OK")
}

// Handle the "PASS" command
func (p *ClientHandler) HandlePASS() {
	if err := p.daddy.driver.CheckUser(p, p.user, p.param); err == nil {
		p.writeMessage(230, "Password ok, continue")
	} else {
		p.writeMessage(530, fmt.Sprintf("Authentication problem: %s", err))
		p.Die()
	}
}
