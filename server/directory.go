package server

import "strconv"
import "bytes"
import "fmt"
import "strings"
import "time"
import "github.com/jehiah/go-strftime"

func (p *Paradise) HandleList() {
	passive := p.lastPassive()
	if passive == nil {
		return
	}
	p.writeMessage(150, "Opening ASCII mode data connection for file list")

	bytes, err := p.dirList()
	if err != nil {
		p.writeMessage(550, err.Error())
	} else {
		if waitTimeout(&passive.waiter, time.Minute) {
			p.writeMessage(550, "Could not get passive connection.")
			p.closePassive(passive)
			return
		}
		if passive.listenFailedAt > 0 {
			p.writeMessage(550, "Could not get passive connection.")
			p.closePassive(passive)
			return
		}
		passive.connection.Write(bytes)
		message := "Closing data connection, sent some bytes"
		p.writeMessage(226, message)
	}
	p.closePassive(passive)
}

func (p *Paradise) dirList() ([]byte, error) {
	var buf bytes.Buffer

	files, err := FileManager.GetFiles(&p.userInfo)
	for _, file := range files {

		if file["isDir"] != "" {
			fmt.Fprintf(&buf, "drw-r--r--")
		} else {
			fmt.Fprintf(&buf, "-rw-r--r--")
		}
		fmt.Fprintf(&buf, " 1 %s %s ", "paradise", "ftp")
		fmt.Fprintf(&buf, lpad(file["size"], 12))
		ts, _ := strconv.ParseInt(file["modTime"], 10, 64)
		fmt.Fprintf(&buf, strftime.Format(" %b %d %H:%M ", time.Unix(ts, 0)))
		fmt.Fprintf(&buf, "%s\r\n", file["name"])
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.Bytes(), err
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
