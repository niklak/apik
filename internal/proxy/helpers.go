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

func writeStatusLine(conn net.Conn, statusCode int, proto string) {
	if _, err := fmt.Fprintf(conn, "%s %d %s\r\n\r\n", proto, statusCode, http.StatusText(statusCode)); err != nil {
		log.Error().Err(err).Msg("Failed to write status line")
	}
}
