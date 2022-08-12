package main

import (
	"io"
	"os"
)

type pipeResult struct {
	Bytes int64
	Error error
}

func singlePipe(from io.Reader, to io.Writer, ch chan pipeResult) {
	n, err := io.Copy(to, from)
	if err != nil && (err != io.EOF || err != os.ErrClosed) {
		ch <- pipeResult{Bytes: n, Error: err}
	} else {
		ch <- pipeResult{Bytes: n}
	}
}

func Pipe(client, server io.ReadWriteCloser) (up, down int64, err error) {
	upch := make(chan pipeResult)
	downch := make(chan pipeResult)
	defer close(upch)
	defer close(downch)

	if client == nil {
		panic("client is nil");
	}
	if server == nil {
		panic("server is nil");
	}

	go singlePipe(client, server, upch)
	go singlePipe(server, client, downch)

	select {
	case u := <-upch:
		up = u.Bytes
		err = u.Error
		client.Close()
		server.Close()
		down = (<-downch).Bytes
	case d := <-downch:
		down = d.Bytes
		err = d.Error
		client.Close()
		server.Close()
		up = (<-upch).Bytes
	}

	return
}
