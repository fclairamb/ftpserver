package server

import "fmt"

// Handle the "USER" command
func (p *ClientHandler) handleUSER() {
	p.user = p.param
	p.writeMessage(331, "OK")
}

// Handle the "PASS" command
func (p *ClientHandler) handlePASS() {
	var err error
	if p.driver, err = p.daddy.driver.AuthUser(p, p.user, p.param); err == nil {
		p.writeMessage(230, "Password ok, continue")
	} else if p.driver == nil {
		p.writeMessage(530, "I can't deal with you (nil driver)")
		p.disconnect()
	} else {
		p.writeMessage(530, fmt.Sprintf("Authentication problem: %s", err))
		p.disconnect()
	}
}
