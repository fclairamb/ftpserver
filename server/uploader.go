package server

import "io"
import "time"

func (self *Paradise) HandleStore() {
	//fmt.Println(self.ip, self.command, self.param)

	self.writeMessage(150, "Data transfer starting")

	passive := self.lastPassive()
	if waitTimeout(&passive.waiter, time.Minute) {
		self.writeMessage(550, "Could not get passive connection.")
		return
	}
	if passive.listenFailedAt > 0 {
		self.writeMessage(550, "Could not get passive connection.")
		return
	}

	_, err := self.storeOrAppend()
	if err == io.EOF {
		self.writeMessage(226, "OK, received some bytes") // TODO send total in message
	} else {
		self.writeMessage(550, "Error with upload: "+err.Error())
	}

	err = passive.connection.Close()
	if err != nil {
		passive.closeFailedAt = time.Now().Unix()
	} else {
		passive.closeSuccessAt = time.Now().Unix()
	}
}

func (self *Paradise) storeOrAppend() (int64, error) {
	var err error
	err = self.readFirst512Bytes()
	if err != nil {
		return 0, err
	}

	// TODO run self.buffer thru mime type checker
	// if mime type bad, reject upload

	// TODO send self.buffer to where u want bits stored

	var total int64
	var n int
	total = int64(len(self.buffer))
	for {
		temp_buffer := make([]byte, 20971520) // reads 20MB at a time
		n, err = self.passiveConn.Read(temp_buffer)
		total += int64(n)

		if err != nil {
			break
		}
		// TODO send temp_buffer to where u want bits stored
		if err != nil {
			break
		}
	}
	//fmt.Println(self.id, " Done ", total)

	return total, err
}

func (self *Paradise) readFirst512Bytes() error {
	self.buffer = make([]byte, 0)
	var err error
	for {
		temp_buffer := make([]byte, 512)
		n, err := self.passiveConn.Read(temp_buffer)

		if err != nil {
			break
		}
		self.buffer = append(self.buffer, temp_buffer[0:n]...)

		if len(self.buffer) >= 512 {
			break
		}
	}

	if err != nil && err != io.EOF {
		return err
	}

	// you have a buffer filled to 512, or less if file is less than 512
	return nil
}
