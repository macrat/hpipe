package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Version      = "1.0.0"
	ProtoName    = "hpipe"
	ProtoVersion = "1"

	helpMode    = flag.Bool("h", false, "Show this help and exit.")
	versionMode = flag.Bool("v", false, "Show version and exit.")
	listen      = flag.String("l", "", "Listen address.")
)

var (
	ErrTimeout = errors.New("timeout")
	ErrConnect = errors.New("failed to establish tunnel")
	ErrProxy   = errors.New("invalid proxy setting")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage:\n  (HTTP->TCP)   hpipe -l [address]:port [host]:port\n  (TCP->HTTP)   hpipe -l [address]:port URL\n  (stdio->HTTP) hpipe URL\n\noptions:\n")
		flag.PrintDefaults()
	}

	zerolog.DurationFieldUnit = time.Second
}

type ReadWriteCloser struct {
	Reader io.Reader
	Writer io.Writer
	Closer io.Closer
}

func (rwc ReadWriteCloser) Read(p []byte) (int, error) {
	return rwc.Reader.Read(p)
}

func (rwc ReadWriteCloser) Write(p []byte) (int, error) {
	return rwc.Writer.Write(p)
}

func (rwc ReadWriteCloser) Close() error {
	if rwc.Closer != nil {
		return rwc.Closer.Close()
	} else {
		return nil
	}
}

type ProcessCloser struct{}

func (pc ProcessCloser) Close() error {
	os.Exit(0)
	return nil
}

func main() {
	flag.Parse()

	switch {
	case *helpMode:
		flag.Usage()
	case *versionMode:
		fmt.Printf("hpipe %s\n", Version)
	case len(flag.Args()) != 1:
		flag.Usage()
		os.Exit(2)
	case *listen != "":
		target := flag.Args()[0]
		u, err := url.Parse(target)
		if err == nil && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ws" || u.Scheme == "wss") {
			err = tcp2http(*listen, u)
		} else {
			err = http2tcp(*listen, target)
		}
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	default:
		stdio := ReadWriteCloser{
			Reader: os.Stdin,
			Writer: os.Stdout,
			Closer: ProcessCloser{},
		}
		if err := stdio2http(stdio, flag.Args()[0]); err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
}
