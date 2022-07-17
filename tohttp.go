package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

type HpipeClient struct {
	Timeout time.Duration
}

func (h HpipeClient) Dial(u *url.URL) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()

	req := (&http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"User-Agent": {"hpipe/" + Version},
			"Upgrade":    {ProtoName + "/" + ProtoVersion},
			"Connection": {"upgrade"},
		},
	}).WithContext(ctx)

	host := u.Host

	proxy, err := http.ProxyFromEnvironment(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrProxy, err)
	} else if proxy != nil {
		host = proxy.Host
		if password, ok := proxy.User.Password(); ok {
			req.SetBasicAuth(proxy.User.Username(), password)
		}
	}

	conn, err := net.Dial("tcp", host)
	if errors.Is(err, context.Canceled) {
		return nil, ErrTimeout
	} else if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnect, err)
	}

	if proxy == nil {
		err = req.Write(conn)
	} else {
		err = req.WriteProxy(conn)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnect, err)
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnect, err)
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("%w: %s", ErrConnect, resp.Status)
	}

	return ReadWriteCloser{
		Reader: reader,
		Writer: conn,
		Closer: conn,
	}, nil
}

func stdio2http(stdio io.ReadWriteCloser, target string) error {
	u, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	dialer := HpipeClient{1 * time.Minute}
	server, err := dialer.Dial(u)
	if err != nil {
		return err
	}

	_, _, err = Pipe(stdio, server)

	return err
}

func tcp2http(listen string, target *url.URL) error {
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}

	dialer := HpipeClient{1 * time.Minute}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		log := log.With().
			Str("remote", conn.RemoteAddr().String()).
			Logger()

		server, err := dialer.Dial(target)
		if err != nil {
			log.Error().Msg("failed to establish connection")
		}

		log.Info().Msg("connection established")

		stime := time.Now()
		up, down, err := Pipe(conn, server)

		log.Info().
			Int64("up_bytes", up).
			Int64("down_bytes", down).
			Dur("duration", time.Since(stime)).
			Err(err).
			Msg("connection closed")
	}
}
