package main

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

func isUpgradable(upgradeHeader string) bool {
	for _, x := range strings.Split(upgradeHeader, ",") {
		xs := strings.SplitN(strings.TrimSpace(x), "/", 2)
		if xs[0] != ProtoName {
			continue
		}
		if len(xs) > 1 && xs[1] != ProtoVersion {
			continue
		}
		return true
	}
	return false
}

type HpipeServer struct {
	Target  string
	Timeout time.Duration
}

func (h HpipeServer) ServeWebsocket(w http.ResponseWriter, r *http.Request) {
	log := log.With().
		Str("proto", "websocket").
		Str("url", r.URL.String()).
		Str("remote", r.RemoteAddr).
		Str("client", r.UserAgent()).
		Logger()

	websocket.Handler(func(conn *websocket.Conn) {
		target, err := net.DialTimeout("tcp", h.Target, h.Timeout)
		var neterr net.Error
		if errors.As(err, &neterr) && neterr.Timeout() {
			log.Error().Err(err).Msg("timeout to connect target")
			http.Error(w, "failed to establish tunnel", http.StatusGatewayTimeout)
			return
		} else if err != nil {
			log.Error().Err(err).Msg("failed to connect target")
			http.Error(w, "failed to establish tunnel", http.StatusBadGateway)
			return
		}
		defer target.Close()

		log.Info().Msg("connection established")

		stime := time.Now()
		up, down, err := Pipe((*WSReadWriteCloser)(conn), target)

		log.Info().
			Int64("up_bytes", up).
			Int64("down_bytes", down).
			Dur("duration", time.Since(stime)).
			Err(err).
			Msg("connection closed")
	}).ServeHTTP(w, r)
}

func (h HpipeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := log.With().
		Str("proto", "hpipe").
		Str("url", r.URL.String()).
		Str("remote", r.RemoteAddr).
		Str("client", r.UserAgent()).
		Logger()

	if r.URL.Path != "/" {
		log.Info().Msg("not found")
		http.NotFound(w, r)
		return
	}

	if r.Header.Get("upgrade") == "" {
		log.Info().Msg("serve working page")
		w.Write([]byte("hpipe is working\n"))
		return
	} else if !isUpgradable(r.Header.Get("upgrade")) {
		h.ServeWebsocket(w, r)
		return
	}

	target, err := net.DialTimeout("tcp", h.Target, h.Timeout)
	var neterr net.Error
	if errors.As(err, &neterr) && neterr.Timeout() {
		log.Error().Err(err).Msg("timeout to connect target")
		http.Error(w, "failed to establish tunnel", http.StatusGatewayTimeout)
		return
	} else if err != nil {
		log.Error().Err(err).Msg("failed to connect target")
		http.Error(w, "failed to establish tunnel", http.StatusBadGateway)
		return
	}
	defer target.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Error().Msg("this server does not support upgrade protocol")
		http.Error(w, "failed to upgrade", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Connection", "upgrade")
	w.Header().Set("Upgrade", ProtoName+"/"+ProtoVersion)
	w.WriteHeader(http.StatusSwitchingProtocols)

	client, reader, err := hj.Hijack()
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade")
		return
	}
	defer client.Close()

	log.Info().Msg("connection established")

	stime := time.Now()
	up, down, err := Pipe(ReadWriteCloser{
		Reader: reader,
		Writer: client,
		Closer: client,
	}, target)

	log.Info().
		Int64("up_bytes", up).
		Int64("down_bytes", down).
		Dur("duration", time.Since(stime)).
		Err(err).
		Msg("connection closed")
}

func http2tcp(listen, target string) error {
	err := http.ListenAndServe(listen, HpipeServer{
		Target:  target,
		Timeout: 1 * time.Minute,
	})
	if err != nil {
		return err
	}
	return nil
}
