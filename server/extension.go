package server

import (
	"paficent/GoFox2X/data"
	"paficent/GoFox2X/protocol"
	"paficent/GoFox2X/transport"
)

type ExtensionHandler func(conn *transport.Conn, params *data.GFSObject)

type ExtensionRouter struct {
	handlers map[string]ExtensionHandler
}

func NewExtensionRouter() *ExtensionRouter {
	return &ExtensionRouter{handlers: make(map[string]ExtensionHandler)}
}

func (r *ExtensionRouter) On(command string, handler ExtensionHandler) *ExtensionRouter {
	r.handlers[command] = handler
	return r
}

func (r *ExtensionRouter) Dispatch(conn *transport.Conn, msg *protocol.Message) bool {
	command, params, ok := protocol.ParseExtensionRequest(msg)
	if !ok {
		return false
	}

	handler, ok := r.handlers[command]
	if !ok {
		return false
	}

	handler(conn, params)
	return true
}
