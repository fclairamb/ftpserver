package server

import "io"
import "time"

func (p *ClientHandler) HandleStore() {
	passive := p.lastPassive()
	if passive == nil {
		return
	}

	p.writeMessage(150, "Data transfer starting")
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

	name := p.Path() + "/" + p.param

	_, err := p.storeOrAppend(passive, name)
	if err == io.EOF {
		p.writeMessage(226, "OK, received some bytes") // TODO send total in message
	} else {
		p.writeMessage(550, "Error with upload: "+err.Error())
	}

	p.closePassive(passive)
}

func (p *ClientHandler) storeOrAppend(passive *Passive, name string) (int64, error) {
	var err error
	/*
	err = p.readFirst512Bytes(passive)
	if err != nil {
		return 0, err
	}
	*/

	file, err := p.daddy.driver.StartFileUpload(p, name)

	if err != nil {
		return 0, err
	}

	// TODO run p.buffer thru mime type checker
	// if mime type bad, reject upload

	// TODO send p.buffer to where u want bits stored

	total := int64(0)
	n := 0
	bytesToRead := 512 // We read 512B and then 4MB
	for {
		temp_buffer := make([]byte, bytesToRead)
		n, err = passive.connection.Read(temp_buffer)
		total += int64(n)

		if err != nil {
			break
		}

		file.Write(temp_buffer[0:n])

		// TODO send temp_buffer to where u want bits stored
		if err != nil {
			break
		}
		bytesToRead = 4 * 1024 * 1024
	}
	file.Close()
	//fmt.Println(p.id, " Done ", total)

	return total, err
}

// This is useless. We could indeed only read 512 bytes the first time around, and it's the driver who should accept
// or not this mimetype.
/*
func (p *ClientHandler) readFirst512Bytes(passive *Passive) error {
	p.buffer = make([]byte, 0)
	var err error
	for {
		temp_buffer := make([]byte, 512)
		n, err := passive.connection.Read(temp_buffer)

		if err != nil {
			break
		}
		p.buffer = append(p.buffer, temp_buffer[0:n]...)

		if len(p.buffer) >= 512 {
			break
		}
	}

	if err != nil && err != io.EOF {
		return err
	}

	// you have a buffer filled to 512, or less if file is less than 512
	return nil
}
*/