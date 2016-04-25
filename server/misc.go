package server

//import "fmt"
import "strings"

func (p *Paradise) HandleSyst() {
	p.writeMessage(215, "UNIX Type: L8")
}

func (p *Paradise) HandlePwd() {
	p.writeMessage(257, "\""+p.path+"\" is the current directory")
}

func (p *Paradise) HandleType() {
	p.writeMessage(200, "Type set to binary")
}

func (p *Paradise) HandleQuit() {
	//fmt.Println("Goodbye")
	p.writeMessage(221, "Goodbye")
	p.theConnection.Close()
	delete(ConnectionMap, p.cid)
}

func (p *Paradise) HandleCwd() {
	if p.param == ".." {
		p.path = "/"
	} else {
		p.path = p.param
	}
	if !strings.HasPrefix(p.path, "/") {
		p.path = "/" + p.path
	}
	p.userInfo["path"] = p.path
	p.writeMessage(250, "CD worked")
}

func (p *Paradise) HandleSize() {
	p.writeMessage(450, "downloads not allowed")
}

func (p *Paradise) HandleStat() {
	p.writeMessage(551, "downloads not allowed")
}
