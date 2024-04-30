package proxy

import (
	"net"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

func HttpProxyConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	logger := log.With().Str("from", r.RemoteAddr).Str("to", r.URL.Host).Logger()

	clientConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	logger.Debug().Msgf("Proxy %s %s", r.Method, r.URL.Host)

	targetConn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		writeStatusLine(clientConn, http.StatusServiceUnavailable, r.Proto)
		return
	}
	defer targetConn.Close()

	writeStatusLine(clientConn, http.StatusOK, r.Proto)

	logger.Debug().Msg("Transfer start")

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go copyIO(wg, targetConn, clientConn)
	go copyIO(wg, clientConn, targetConn)
	wg.Wait()

	logger.Debug().Msg("Transfer complete")
}
