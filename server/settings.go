package server

import "github.com/naoina/toml"
import "os"
import "io/ioutil"

type ParadiseSettings struct {
	Host           string
	Port           int
	MaxConnections int
	MaxPassive     int
	Exec           string
}

func ReadSettings() ParadiseSettings {
	f, err := os.Open("conf/settings.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var config ParadiseSettings
	if err := toml.Unmarshal(buf, &config); err != nil {
		panic(err)
	}
	return config
}
