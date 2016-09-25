package server

import "fmt"
import "time"
import "net/http"

func countdown(upsince int64) string {
	secs := time.Now().Unix() - upsince
	us := time.Unix(secs, 0)
	str := us.UTC().String()
	return str[11:19]
}

func trimGuid(guid string) string {
	return guid[0:6]
}

func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "%d client(s), %d passive(s), Up for %s\n",
		len(ConnectionMap), PassiveCount, countdown(UpSince))

	for k, v := range ConnectionMap {
		fmt.Fprintf(w, "   %s %s, %s\n", trimGuid(k), countdown(v.connectedAt), v.user)
		for pk, pv := range v.passives {
			fmt.Fprintf(w, "     %s %s, %d %s %s\n", trimGuid(pk), countdown(pv.listenAt), pv.port, pv.command, pv.param)
		}
	}
}

func Monitor() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":5010", nil)
}
