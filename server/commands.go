package server

func MakeCommandMap() map[string]func(*Paradise) {
	m := make(map[string]func(*Paradise))

	m["USER"] = (*Paradise).HandleUser
	m["PASS"] = (*Paradise).HandlePass
	m["STOR"] = (*Paradise).HandleStore
	m["APPE"] = (*Paradise).HandleStore
	m["STAT"] = (*Paradise).HandleStat

	m["SYST"] = (*Paradise).HandleSyst
	m["PWD"] = (*Paradise).HandlePwd
	m["TYPE"] = (*Paradise).HandleType
	m["PASV"] = (*Paradise).HandlePassive
	m["EPSV"] = (*Paradise).HandlePassive
	m["NLST"] = (*Paradise).HandleList
	m["LIST"] = (*Paradise).HandleList
	m["QUIT"] = (*Paradise).HandleQuit
	m["CWD"] = (*Paradise).HandleCwd
	m["SIZE"] = (*Paradise).HandleSize
	m["RETR"] = (*Paradise).HandleRetr

	return m
}
