package module

import (
	"github.com/liasece/micserver/connect"
	"github.com/liasece/micserver/msg"
	"github.com/liasece/micserver/util"
	"net"
)

type clientEventHandler struct {
	mod *BaseModule

	fonNewClient     func(client *connect.Client)
	fonRecvMsg       func(client *connect.Client, msgbin *msg.MessageBinary)
	fonAcceptConnect func(conn net.Conn)
}

func (this *clientEventHandler) RegOnNewClient(
	cb func(client *connect.Client)) {
	this.fonNewClient = cb
}

func (this *clientEventHandler) OnNewClient(client *connect.Client) {
	servertype := util.GetServerIDType(this.mod.ModuleID)
	client.SetBindServer(servertype, this.mod.ModuleID)

	if this.fonNewClient != nil {
		this.fonNewClient(client)
	}
}

func (this *clientEventHandler) RegOnRecvMsg(
	cb func(client *connect.Client, msgbin *msg.MessageBinary)) {
	this.fonRecvMsg = cb
}

func (this *clientEventHandler) OnRecvMsg(
	client *connect.Client, msgbin *msg.MessageBinary) {
	if this.fonRecvMsg != nil {
		this.fonRecvMsg(client, msgbin)
	}
}

func (this *clientEventHandler) RegOnAcceptConnect(cb func(conn net.Conn)) {
	this.fonAcceptConnect = cb
}

func (this *clientEventHandler) OnAcceptConnect(conn net.Conn) {
	if this.fonAcceptConnect != nil {
		this.fonAcceptConnect(conn)
	}
}
