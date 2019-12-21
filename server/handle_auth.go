// Package server provides all the tools to build your own FTP server: The core library and the driver.
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
	c.driver, err = c.server.driver.AuthUser(c, c.user, c.param)

	switch {
	case err == nil:
		c.writeMessage(StatusUserLoggedIn, "Password ok, continue")
	case err != nil:
		c.writeMessage(StatusNotLoggedIn, fmt.Sprintf("Authentication problem: %v", err))
		c.disconnect()
	default:
		c.writeMessage(StatusNotLoggedIn, "I can't deal with you (nil driver)")
		c.disconnect()
	}

	return nil
}
