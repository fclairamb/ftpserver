package server

import "fmt"

// Handle the "USER" command
func (c *clientHandler) handleUSER() error {
	c.user = c.param
	c.writeMessage(StatusUserOK, "OK")
	return nil
}

// Handle the "PASS" command
func (c *clientHandler) handlePASS() error {
	var err error
	if c.driver, err = c.server.driver.AuthUser(c, c.user, c.param); err == nil {
		c.writeMessage(StatusUserLoggedIn, "Password ok, continue")
	} else if err != nil {
		c.writeMessage(StatusNotLoggedIn, fmt.Sprintf("Authentication problem: %v", err))
		c.disconnect()
	} else {
		c.writeMessage(StatusNotLoggedIn, "I can't deal with you (nil driver)")
		c.disconnect()
	}
	return nil
}
