package server

import "fmt"
import "time"
import (
	"net/http"
	"gopkg.in/inconshreveable/log15.v2"
)

func oneSPrecision(d time.Duration) time.Duration {
	return d - (d % time.Second)
}

func (server *FtpServer) handler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()

	fmt.Fprintf(w,
		"%d client(s), Up for %s\n",
		len(server.connectionsById),
		oneSPrecision(now.Sub(server.StartTime)),
	)

	server.connectionsMutex.RLock()
	defer server.connectionsMutex.RUnlock()

	for k, v := range server.connectionsById {
		fmt.Fprintf(w, "   %d %s, %s\n", k, oneSPrecision(now.Sub(v.connectedAt)), v.user)
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
