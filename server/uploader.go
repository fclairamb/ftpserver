package server

import "io"
import (
	"time"
	"os"
	"fmt"
	"gopkg.in/inconshreveable/log15.v2"
)

func (p *ClientHandler) HandleStore() {
	p.handleStoreAndAppend(false)
}

func (p *ClientHandler) HandleAppend() {
	p.handleStoreAndAppend(true)
}

func (p *ClientHandler) handleStoreAndAppend(append bool) {
	passive := p.lastPassive()
	if passive == nil {
		return
	}
	defer p.closePassive(passive)

	p.writeMessage(150, "Data transfer starting")
	if waitTimeout(&passive.waiter, time.Minute) {
		p.writeMessage(550, "Could not get passive connection.")
		return
	}
	if passive.listenFailedAt > 0 {
		p.writeMessage(550, "Could not get passive connection.")
		return
	}

	name := p.Path() + "/" + p.param


	if total, err := p.storeOrAppend(passive, append, name); err == nil {
		p.writeMessage(226, fmt.Sprintf("OK, received %d bytes", total))
	} else {
		p.writeMessage(550, "Error with upload: "+err.Error())
	}
}

func (p *ClientHandler) storeOrAppend(passive *Passive, append bool, name string) (int64, error) {
	var err error

	flag := 0

	if append {
		flag |= os.O_APPEND
	}

	file, err := p.daddy.driver.StartFileUpload(p, name, flag)

	if err != nil {
		return 0, err
	}
	defer file.Close()

	/*
	// This doesn't work if we upload a smaller file than the original one. The append has to be dealt with while the file is opened.
	if append {
		// Let's get it at the end
		file.Seek(0, 2)
	}
	*/

	total := int64(0)
	n := 0
	bytesToRead := 512 // We read 512B and then 4MB
	for {
		temp_buffer := make([]byte, bytesToRead)
		n, err = passive.connection.Read(temp_buffer)
		total += int64(n)

		if err != nil {
			log15.Error("Error while reading", "err", err)
			break
		}

		_, err := file.Write(temp_buffer[0:n])

		if err != nil {
			log15.Error("Error while writing", "err", err)
			break
		}

		bytesToRead = 4 * 1024 * 1024
	}

	if err == io.EOF {
		return total, nil
	} else {
		return total, err
	}
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