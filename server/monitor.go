package server

import "fmt"
import "time"
import (
	"net/http"
	"gopkg.in/inconshreveable/log15.v2"
)

func countdown(upsince int64) string {
	secs := time.Now().Unix() - upsince
	us := time.Unix(secs, 0)
	str := us.UTC().String()
	return str[11:19]
}

func trimGuid(guid string) string {
	return guid[0:6]
}

func (server *FtpServer) handler(w http.ResponseWriter, r *http.Request) {
	server.connectionsMutex.RLock()
	defer server.connectionsMutex.RUnlock()

	fmt.Fprintf(w, "%d client(s), %d passive(s), Up for %s\n",
		len(server.ConnectionsById), server.PassiveCount, countdown(server.StartTime))

	for k, v := range server.ConnectionsById {
		fmt.Fprintf(w, "   %s %s, %s\n", trimGuid(k), countdown(v.connectedAt), v.user)
		for pk, pv := range v.passives {
			fmt.Fprintf(w, "     %s %s, %d %s %s\n", trimGuid(pk), countdown(pv.listenAt), pv.port, pv.command, pv.param)
		}
	}
}

func (server *FtpServer) handlerStop(w http.ResponseWriter, r *http.Request) {
	server.Listener.Close()
}

func (server *FtpServer) Monitor() error {
	http.HandleFunc("/", server.handler)
	http.HandleFunc("/stop", server.handlerStop)

	lstAddr := fmt.Sprintf(":%d", server.Settings.MonitorPort)

	log15.Info("Monitor listening", "addr", lstAddr)
	return http.ListenAndServe(lstAddr, nil)
}
