package server

import (
	"bytes"
	"fmt"
	"github.com/jehiah/go-strftime"
	"strconv"
	"strings"
	"time"
)

func (self *Paradise) HandleList() {
	//fmt.Println(self.ip, self.command, self.param)

	self.writeMessage(150, "Opening ASCII mode data connection for file list")

	bytes, err := self.dirList()
	if err != nil {
		self.writeMessage(550, err.Error())
	} else {
		if waitTimeout(&self.waiter, time.Minute) {
			self.writeMessage(550, "Could not get passive connection.")
			return
		}
		if self.passiveListenFailedAt > 0 {
			self.writeMessage(550, "Could not get passive connection.")
			return
		}
		self.passiveConn.Write(bytes)
		message := "Closing data connection, sent some bytes"
		self.writeMessage(226, message)

		err := self.passiveConn.Close()
		if err != nil {
			self.passiveCloseFailedAt = time.Now().Unix()
		} else {
			self.passiveCloseSuccessAt = time.Now().Unix()
		}
	}
}

func (self *Paradise) dirList() ([]byte, error) {
	var buf bytes.Buffer

	files := []int{1, 2, 3, 4, 5} // change to real list of files
	for _, _ = range files {

		if false { // change to really test for isDir
			fmt.Fprintf(&buf, "drw-r--r--")
		} else {
			fmt.Fprintf(&buf, "-rw-r--r--")
		}
		fmt.Fprintf(&buf, " 1 %s %s ", "paradise", "ftp")
		fmt.Fprintf(&buf, lpad(strconv.Itoa(13984), 12))                // change to real file size
		fmt.Fprintf(&buf, strftime.Format(" %b %d %H:%M ", time.Now())) // change to real file date
		fmt.Fprintf(&buf, "%s\r\n", "paradise.txt")                     // change to real filename
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.Bytes(), nil
}

func lpad(input string, length int) (result string) {
	if len(input) < length {
		result = strings.Repeat(" ", length-len(input)) + input
	} else if len(input) == length {
		result = input
	} else {
		result = input[0:length]
	}
	return
}
