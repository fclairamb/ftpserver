package server

import (
	"bytes"
	"fmt"
	"github.com/jehiah/go-strftime"
	"strconv"
	"strings"
	"time"
)

func (p *Paradise) HandleList() {
	//fmt.Println(p.ip, p.command, p.param)

	p.writeMessage(150, "Opening ASCII mode data connection for file list")

	bytes, err := p.dirList()
	if err != nil {
		p.writeMessage(550, err.Error())
	} else {
		passive := p.lastPassive()
		if waitTimeout(&passive.waiter, time.Minute) {
			p.writeMessage(550, "Could not get passive connection.")
			return
		}
		if passive.listenFailedAt > 0 {
			p.writeMessage(550, "Could not get passive connection.")
			return
		}
		passive.connection.Write(bytes)
		message := "Closing data connection, sent some bytes"
		p.writeMessage(226, message)

		err := passive.connection.Close()
		if err != nil {
			passive.closeFailedAt = time.Now().Unix()
		} else {
			passive.closeSuccessAt = time.Now().Unix()
		}
	}
}

func (p *Paradise) dirList() ([]byte, error) {
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
