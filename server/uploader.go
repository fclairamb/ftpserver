package server

import "io"
import "time"

func (p *Paradise) HandleStore() {
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

	_, err := p.storeOrAppend(passive)
	if err == io.EOF {
		p.writeMessage(226, "OK, received some bytes") // TODO send total in message
	} else {
		p.writeMessage(550, "Error with upload: "+err.Error())
	}

	p.closePassive(passive)
}

func (p *Paradise) storeOrAppend(passive *Passive) (int64, error) {
	var err error
	err = p.readFirst512Bytes(passive)
	if err != nil {
		return 0, err
	}

	// TODO run p.buffer thru mime type checker
	// if mime type bad, reject upload

	// TODO send p.buffer to where u want bits stored

	var total int64
	var n int
	total = int64(len(p.buffer))
	for {
		temp_buffer := make([]byte, 20971520) // reads 20MB at a time
		n, err = passive.connection.Read(temp_buffer)
		total += int64(n)

		if err != nil {
			break
		}
		// TODO send temp_buffer to where u want bits stored
		if err != nil {
			break
		}
	}
	//fmt.Println(p.id, " Done ", total)

	return total, err
}

func (p *Paradise) readFirst512Bytes(passive *Passive) error {
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
