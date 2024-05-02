package request

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// TraceTimings represents the timings of the request
type TraceTimings struct {
	ConnGet           time.Time
	ConnGot           time.Time
	PutIdleConn       time.Time
	DNSStart          time.Time
	DNSDone           time.Time
	ConnectStart      time.Time
	ConnectDone       time.Time
	TLSHandshakeStart time.Time
	TLSHandshakeDone  time.Time
	FirstByte         time.Time
	Wait100Continue   time.Time
	WroteHeaders      time.Time
	WroteRequest      time.Time
}

// TraceConnect represents a connection trace
type TraceConnect struct {
	Network string
	Address string
	Error   error
	Time    time.Time
}

// TLSHandshake represents a TLS handshake
type TLSHandshake struct {
	State tls.ConnectionState
	Error error
}

// TraceInfo represents the trace information
type TraceInfo struct {
	Timings      TraceTimings
	GetConnHost  string
	GotConn      httptrace.GotConnInfo
	DNSDone      httptrace.DNSDoneInfo
	DNSStart     httptrace.DNSStartInfo
	WroteRequest httptrace.WroteRequestInfo
	PutIdleError error
	ConnectStart []TraceConnect
	ConnectDone  []TraceConnect
}

// Hooks returns a httptrace.ClientTrace with the trace hooks
func (s *TraceInfo) hooks() *httptrace.ClientTrace {
	t := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			s.Timings.ConnGet = time.Now()
			s.GetConnHost = hostPort
		},
		GotConn: func(info httptrace.GotConnInfo) {
			s.Timings.ConnGot = time.Now()
			s.GotConn = info
		},
		PutIdleConn: func(err error) {
			s.Timings.PutIdleConn = time.Now()
			s.PutIdleError = err
		},
		GotFirstResponseByte: func() {
			s.Timings.FirstByte = time.Now()
		},
		Got100Continue: func() {
			s.Timings.Wait100Continue = time.Now()
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			s.Timings.DNSStart = time.Now()
			s.DNSStart = info
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			s.Timings.DNSDone = time.Now()
			s.DNSDone = info
		},
		ConnectStart: func(network, addr string) {

			s.ConnectStart = append(s.ConnectStart, TraceConnect{
				Network: network,
				Address: addr,
				Time:    time.Now(),
			})
			if s.Timings.ConnectStart.IsZero() {
				s.Timings.ConnectStart = time.Now()
			}
		},
		ConnectDone: func(network, addr string, err error) {
			s.ConnectDone = append(s.ConnectDone, TraceConnect{
				Network: network,
				Address: addr,
				Error:   err,
				Time:    time.Now(),
			})
			s.Timings.ConnectDone = time.Now()
		},
		TLSHandshakeStart: func() {
			s.Timings.TLSHandshakeStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			s.Timings.TLSHandshakeDone = time.Now()
		},
		Wait100Continue: func() {
			s.Timings.Wait100Continue = time.Now()
		},
		WroteHeaders: func() {
			s.Timings.WroteHeaders = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			s.Timings.WroteRequest = time.Now()
		},
	}
	return t
}

// createTraceContext creates a new context with a trace hook
func createTraceContext(ctx context.Context) (info *TraceInfo, traceCtx context.Context) {

	info = &TraceInfo{}
	traceCtx = httptrace.WithClientTrace(ctx, info.hooks())
	return
}
