package server

import "fmt"

// Handle the "USER" command
func (c *clientHandler) handleUSER() {
	c.user = c.param
	c.writeMessage(331, "OK")
}

// Handle the "PASS" command
func (c *clientHandler) handlePASS() {
	var err error
	if c.driver, err = c.daddy.driver.AuthUser(c, c.user, c.param); err == nil {
		c.writeMessage(230, "Password ok, continue")
	} else if err != nil {
		c.writeMessage(530, fmt.Sprintf("Authentication problem: %v", err))
		c.disconnect()
	} else {
		c.writeMessage(530, "I can't deal with you (nil driver)")
		c.disconnect()
	}
}
