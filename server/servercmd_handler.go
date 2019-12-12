package server

import (
	"github.com/liasece/micserver/connect"
	"github.com/liasece/micserver/msg"
	"github.com/liasece/micserver/server/base"
	"github.com/liasece/micserver/servercomm"
	"github.com/liasece/micserver/session"
)

type serverCmdHandler struct {
	server *Server

	serverHook base.ServerHook
}

func (this *serverCmdHandler) HookServer(serverHook base.ServerHook) {
	this.serverHook = serverHook
}

func (this *serverCmdHandler) onForwardToModule(conn *connect.Server,
	smsg *servercomm.SForwardToModule) {
	if this.serverHook != nil {
		msg := &servercomm.ModuleMessage{
			FromModule: conn.ModuleInfo,
			MsgID:      smsg.MsgID,
			Data:       smsg.Data,
		}
		this.serverHook.OnModuleMessage(msg)
	}
}

func (this *serverCmdHandler) onForwardFromGate(conn *connect.Server,
	smsg *servercomm.SForwardFromGate) {
	if this.serverHook != nil {
		msg := &servercomm.ClientMessage{
			FromModule:   conn.ModuleInfo,
			ClientConnID: smsg.ClientConnID,
			MsgID:        smsg.MsgID,
			Data:         smsg.Data,
		}
		uuid := session.GetUUIDFromMap(smsg.Session)
		var se *session.Session
		if uuid != "" {
			se = this.server.GetSession(uuid)
		}
		if se == nil {
			se = session.NewSessionFromMap(smsg.Session)
		}
		this.serverHook.OnClientMessage(se, msg)
	}
}

func (this *serverCmdHandler) onForwardToClient(smsg *servercomm.SForwardToClient) {
	err := this.server.DoSendBytesToClient(smsg.FromModuleID, smsg.ToGateID,
		smsg.ToClientID, smsg.MsgID, smsg.Data)
	if err != nil {
		if err == ErrTargetClientDontExist {
			this.server.Debug("this.doSendBytesToClient Err:%s ServerMsg:%+v",
				err.Error(), smsg)
		} else {
			this.server.Error("this.doSendBytesToClient Err:%s ServerMsg:%+v",
				err.Error(), smsg)
		}
	}
}

func (this *serverCmdHandler) onUpdateSession(smsg *servercomm.SUpdateSession) {
	var connectedSession *session.Session
	if this.server.gateBase != nil {
		client := this.server.GetClient(smsg.ClientConnID)
		if client != nil {
			client.Session.FromMap(smsg.Session)
			connectedSession = client.Session
			// if client.Session.GetUUID() != "" {
			// 	this.server.Info("[gate] 用户登陆成功 %s", smsg.GetJson())
			// }
		}
	}

	// 尝试更新本地 session
	if smsg.SessionUUID != "" {
		// 先从连接中的session复制
		s := connectedSession
		localsession := this.server.sessionManager.GetSession(smsg.SessionUUID)
		if localsession != nil {
			if connectedSession != nil && connectedSession != localsession {
				// 不是同一个session对象，需要将本地session复制为最新链接的session
				connectedSession.OnlyAddKeyFromSession(localsession)
				this.server.sessionManager.Store(connectedSession)
				localsession = connectedSession
			}
			s = localsession
		}
		if s == nil {
			s = &session.Session{}
			this.server.sessionManager.UpdateSessionUUID(smsg.SessionUUID, s)
		}
		this.server.sessionManager.MustUpdateFromMap(s, smsg.Session)
		this.server.Debug("[onUpdateSession] Session Manager Update: %+v From:%s To:%s",
			smsg.Session, smsg.FromModuleID, smsg.ToModuleID)
	}
}

// 当一个服务器成功加入网络时调用
func (this *serverCmdHandler) OnServerJoinSubnet(server *connect.Server) {
	this.server.onServerJoinSubnet(server)
}

func (this *serverCmdHandler) OnRecvSubnetMsg(conn *connect.Server,
	msgbinary *msg.MessageBinary) {
	switch msgbinary.CmdID {
	case servercomm.SForwardToModuleID:
		// 服务器间用户空间消息转发
		if this.serverHook != nil {
			layerMsg := &servercomm.SForwardToModule{}
			layerMsg.ReadBinary(msgbinary.ProtoData)
			this.onForwardToModule(conn, layerMsg)
		}
	case servercomm.SForwardFromGateID:
		// Gateway 转发过来的客户端消息
		layerMsg := &servercomm.SForwardFromGate{}
		layerMsg.ReadBinary(msgbinary.ProtoData)
		this.onForwardFromGate(conn, layerMsg)
	case servercomm.SForwardToClientID:
		// 其他服务器转发过来的，要发送到客户端的消息
		var layerMsg *servercomm.SForwardToClient
		if obj := msgbinary.GetObj(); obj != nil {
			if m, ok := obj.(*servercomm.SForwardToClient); ok {
				layerMsg = m
			}
		}
		if layerMsg == nil {
			layerMsg = &servercomm.SForwardToClient{}
			layerMsg.ReadBinary(msgbinary.ProtoData)
		}
		this.onForwardToClient(layerMsg)
	case servercomm.SUpdateSessionID:
		// 客户端会话更新
		layerMsg := &servercomm.SUpdateSession{}
		layerMsg.ReadBinary(msgbinary.ProtoData)
		this.onUpdateSession(layerMsg)
	case servercomm.SStartMyNotifyCommandID:
	case servercomm.SROCBindID:
		// ROC 对象绑定
		layerMsg := &servercomm.SROCBind{}
		layerMsg.ReadBinary(msgbinary.ProtoData)
		this.server.ROCServer.onMsgROCBind(layerMsg)
	case servercomm.SROCRequestID:
		// ROC 调用请求
		layerMsg := &servercomm.SROCRequest{}
		layerMsg.ReadBinary(msgbinary.ProtoData)
		this.server.ROCServer.onMsgROCRequest(layerMsg)
	case servercomm.SROCResponseID:
		// ROC 调用返回
		layerMsg := &servercomm.SROCResponse{}
		layerMsg.ReadBinary(msgbinary.ProtoData)
		this.server.ROCServer.onMsgROCResponse(layerMsg)
	default:
		msgid := msgbinary.CmdID
		msgname := servercomm.MsgIdToString(msgid)
		this.server.Error("[SubnetManager.OnRecvTCPMsg] 未知消息 %d:%s",
			msgid, msgname)
	}
}
