package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

func ProxyConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	logger := log.With().Str("from", r.RemoteAddr).Str("to", r.URL.Host).Logger()

	logger.Debug().Msg("Proxy request")

	clientConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	targetConn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		writeStatusLine(clientConn, http.StatusServiceUnavailable, r)
		return
	}
	defer targetConn.Close()

	writeStatusLine(clientConn, http.StatusOK, r)

	logger.Debug().Msg("Transfer start")

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go copyIO(wg, targetConn, clientConn)
	go copyIO(wg, clientConn, targetConn)
	wg.Wait()

	logger.Debug().Msg("Transfer complete")
}

func writeStatusLine(conn net.Conn, statusCode int, r *http.Request) {
	if _, err := fmt.Fprintf(conn, "HTTP/%d.%d %d %s\r\n\r\n", r.ProtoMajor,
		r.ProtoMinor, statusCode, http.StatusText(statusCode)); err != nil {
		log.Error().Err(err).Msg("Failed to write status line")
	}
}

func copyIO(wg *sync.WaitGroup, dst, src net.Conn) {
	defer wg.Done()
	if _, err := io.Copy(dst, src); err != nil {
		if !strings.Contains(err.Error(), "use of closed network connection") {
			log.Error().Err(err).Msg("")
		}
		return
	}
	dst.Close()

}
