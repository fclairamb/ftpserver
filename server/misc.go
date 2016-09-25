package server

//import "fmt"
import "strings"

func (p *ClientHandler) HandleSyst() {
	p.writeMessage(215, "UNIX Type: L8")
}

func (p *ClientHandler) HandlePwd() {
	p.writeMessage(257, "\""+p.path+"\" is the current directory")
}

func (p *ClientHandler) HandleType() {
	p.writeMessage(200, "Type set to binary")
}

func (p *ClientHandler) HandleQuit() {
	//fmt.Println("Goodbye")
	p.writeMessage(221, "Goodbye")
	p.conn.Close()
	delete(p.daddy.ConnectionMap, p.cid)
}

func (p *ClientHandler) HandleCwd() {
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

func (p *ClientHandler) HandleSize() {
	p.writeMessage(450, "downloads not allowed")
}

func (p *ClientHandler) HandleStat() {
	p.writeMessage(551, "downloads not allowed")
}
