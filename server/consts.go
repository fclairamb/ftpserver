// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

// from @stevenh's PR proposal
// https://github.com/fclairamb/ftpserver/blob/becc125a0770e3b670c4ced7e7bd12594fb024ff/server/consts.go

// Status codes as documented by:
// https://tools.ietf.org/html/rfc959
// https://tools.ietf.org/html/rfc2428
// https://tools.ietf.org/html/rfc2228
const (
	// 100 Series - The requested action is being initiated, expect another reply before
	// proceeding with a new command.
	StatusFileStatusOK = 150 // RFC 959, 4.2.1

	// 200 Series - The requested action has been successfully completed.
	StatusOK                 = 200 // RFC 959, 4.2.1
	StatusNotImplemented     = 202 // RFC 959, 4.2.1
	StatusSystemStatus       = 211 // RFC 959, 4.2.1
	StatusDirectoryStatus    = 212 // RFC 959, 4.2.1
	StatusFileStatus         = 213 // RFC 959, 4.2.1
	StatusHelpMessage        = 214 // RFC 959, 4.2.1
	StatusSystemType         = 215 // RFC 959, 4.2.1
	StatusServiceReady       = 220 // RFC 959, 4.2.1
	StatusClosingControlConn = 221 // RFC 959, 4.2.1
	StatusClosingDataConn    = 226 // RFC 959, 4.2.1
	StatusEnteringPASV       = 227 // RFC 959, 4.2.1
	StatusEnteringEPSV       = 229 // RFC 2428, 3
	StatusUserLoggedIn       = 230 // RFC 959, 4.2.1
	StatusAuthAccepted       = 234 // RFC 2228, 3
	StatusFileOK             = 250 // RFC 959, 4.2.1
	StatusPathCreated        = 257 // RFC 959, 4.2.1

	// 300 Series - The command has been accepted, but the requested action is on hold,
	// pending receipt of further information.
	StatusUserOK            = 331 // RFC 959, 4.2.1
	StatusFileActionPending = 350 // RFC 959, 4.2.1

	// 400 Series - The command was not accepted and the requested action did not take place,
	// but the error condition is temporary and the action may be requested again.
	StatusServiceNotAvailable = 421 // RFC 959, 4.2.1
	StatusFileActionNotTaken  = 450 // RFC 959, 4.2.1

	// 500 Series - Syntax error, command unrecognized and the requested action did not take
	// place. This may include errors such as command line too long.
	StatusSyntaxErrorNotRecognised = 500 // RFC 959, 4.2.1
	StatusSyntaxErrorParameters    = 501 // RFC 959, 4.2.1
	StatusCommandNotImplemented    = 502 // RFC 959, 4.2.1
	StatusNotLoggedIn              = 530 // RFC 959, 4.2.1
	StatusActionNotTaken           = 550 // RFC 959, 4.2.1
)
