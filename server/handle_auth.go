package server

import "fmt"

// Handle the "USER" command
func (p *clientHandler) handleUSER() {
	p.user = p.param
	p.writeMessage(331, "OK")
}

// Handle the "PASS" command
func (p *clientHandler) handlePASS() {
	var err error
	if p.driver, err = p.daddy.driver.AuthUser(p, p.user, p.param); err == nil {
		p.writeMessage(230, "Password ok, continue")
	} else if err != nil {
		p.writeMessage(530, fmt.Sprintf("Authentication problem: %v", err))
		p.disconnect()
	} else {
		p.writeMessage(530, "I can't deal with you (nil driver)")
		p.disconnect()
	}
}
