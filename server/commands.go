package server

func MakeCommandMap() map[string]string {
	m := make(map[string]string)

	m["USER"] = "HandleUser"
	m["PASS"] = "HandlePass"
	m["STOR"] = "HandleStore"
	m["APPE"] = "HandleStore"
	m["STAT"] = "HandleStat"

	m["SYST"] = "HandleSyst"
	m["PWD"] = "HandlePwd"
	m["TYPE"] = "HandleType"
	m["PASV"] = "HandlePassive"
	m["EPSV"] = "HandlePassive"
	m["NLST"] = "HandleList"
	m["LIST"] = "HandleList"
	m["QUIT"] = "HandleQuit"
	m["CWD"] = "HandleCwd"
	m["SIZE"] = "HandleSize"
	m["RETR"] = "HandleRetr"

	return m
}
