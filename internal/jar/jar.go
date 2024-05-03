package jar

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/publicsuffix"
)

type DebugJar struct {
	inner  *cookiejar.Jar
	logger zerolog.Logger
}

func (j *DebugJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.logger.Debug().Str("url", u.String()).Interface("cookies", cookies).Msg("Setting cookies")
	j.inner.SetCookies(u, cookies)
}

func (j *DebugJar) Cookies(u *url.URL) []*http.Cookie {
	cookies := j.inner.Cookies(u)
	j.logger.Debug().Str("url", u.String()).Interface("cookies", cookies).Msg("Getting cookies")
	return cookies
}

func New() *DebugJar {
	inner, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	logger := log.With().Str("module", "jar").Str("component", "Jar").Logger()
	return &DebugJar{inner: inner, logger: logger}

}
