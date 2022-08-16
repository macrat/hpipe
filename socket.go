package main

import (
	"golang.org/x/net/websocket"
)

type WSReadWriteCloser websocket.Conn

func (ws *WSReadWriteCloser) Read(buf []byte) (int, error) {
	var b []byte
	if err := websocket.Message.Receive((*websocket.Conn)(ws), &b); err != nil {
		return 0, err
	}
	return copy(buf, b), nil
}

func (ws *WSReadWriteCloser) Write(buf []byte) (int, error) {
	return len(buf), websocket.Message.Send((*websocket.Conn)(ws), buf)
}

func (ws *WSReadWriteCloser) Close() error {
	return (*websocket.Conn)(ws).Close()
}
