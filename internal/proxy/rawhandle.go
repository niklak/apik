package proxy

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type requestLine struct {
	method    string
	authority string
	proto     string
}

func readRequestLine(scanner *bufio.Scanner) (rLine *requestLine, err error) {

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		rLine = &requestLine{
			method:    parts[0],
			authority: parts[1],
			proto:     parts[2],
		}
		break
	}
	err = scanner.Err()
	return
}

func ProxyConnectHandle(clientConn net.Conn) {

	scanner := bufio.NewScanner(clientConn)

	r, err := readRequestLine(scanner)

	if err != nil {
		writeStatusLine(clientConn, http.StatusInternalServerError, r.proto)
		clientConn.Close()
		return
	}

	if r.method != http.MethodConnect {
		writeStatusLine(clientConn, http.StatusMethodNotAllowed, r.proto)
		clientConn.Close()
		return
	}

	logger := log.With().
		Str("from", clientConn.RemoteAddr().String()).
		Str("to", r.authority).
		Logger()

	targetConn, err := net.Dial("tcp", r.authority)
	if err != nil {
		writeStatusLine(clientConn, http.StatusInternalServerError, r.proto)
		clientConn.Close()
		return
	}
	defer targetConn.Close()

	writeStatusLine(clientConn, http.StatusOK, r.proto)

	logger.Debug().Msg("Transfer start")
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go copyIO(wg, targetConn, clientConn)
	go copyIO(wg, clientConn, targetConn)
	wg.Wait()

	fmt.Printf("Transfer complete\n")
}
