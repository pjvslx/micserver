package subnet

import (
	"fmt"
	"time"

	"github.com/liasece/micserver/connect"
	"github.com/liasece/micserver/msg"
	"github.com/liasece/micserver/servercomm"
	"github.com/liasece/micserver/util/hash"
	"github.com/liasece/micserver/util/monitor"
	"github.com/liasece/micserver/util/sysutil"
)

// 服务器消息处理封包
type ConnectMsgQueueStruct struct {
	conn *connect.Server
	msg  *msg.MessageBinary
}

// 当TCP连接被移除时调用
func (this *SubnetManager) onConnectClose(conn *connect.Server) {
	this.RemoveServer(conn.GetTempID())
}

// 当收到TCP消息时调用
func (this *SubnetManager) OnRecvTCPMsg(conn *connect.Server,
	msgbinary *msg.MessageBinary) {
	if this.subnetHook != nil {
		this.subnetHook.OnRecvSubnetMsg(conn, msgbinary)
	} else {
		this.Syslog("this.SubnetCallback.fonRecvMsg == nil MsgID[%d]",
			msgbinary.GetMsgID())
	}
}

// 获取TCP消息的消息处理通道
func (this *SubnetManager) getRecvTCPMsgParseChan(conn *connect.Server,
	maxChan int32, msgbinary *msg.MessageBinary) int32 {
	chankey := ""
	if conn.ModuleInfo != nil {
		chankey = conn.ModuleInfo.ModuleID
	}
	switch msgbinary.GetMsgID() {
	case servercomm.SForwardToClientID:
		layerMsg := &servercomm.SForwardToClient{}
		layerMsg.ReadBinary(msgbinary.ProtoData)
		msgbinary.SetObj(layerMsg)
		chankey = layerMsg.ToClientID
	}
	if chankey != "" {
		hash := int32(hash.GetStringHash(chankey))
		if hash < 0 {
			hash = -hash
		}
		return hash % maxChan
	}
	return 0
}

// 当新增一个服务器连接时调用
func (this *SubnetManager) OnCreateNewServer(conn *connect.Server) {
}

// 当收到了一个服务器消息时调用
func (this *SubnetManager) onConnectRecv(conn *connect.Server,
	msgbin *msg.MessageBinary) {
	if conn.GetSCType() == connect.ServerSCTypeTask {
		curtime := time.Now().Unix()
		if conn.IsTerminateTimeout(curtime) {
			this.onClientDisconnected(conn)
			this.Error("[SubnetManager.handleConnection] "+
				"长时间未通过验证，断开连接 TmpID[%d]",
				conn.GetTempID())
			return
		}
		if conn.IsTerminateForce() {
			this.onClientDisconnected(conn)
			this.Syslog("[SubnetManager.handleConnection] "+
				"服务器主动断开连接 TmpID[%s]", conn.GetTempID())
			return
		}
	}
	switch msgbin.GetMsgID() {
	case servercomm.STestCommandID:
		recvmsg := &servercomm.STestCommand{}
		recvmsg.ReadBinary([]byte(msgbin.ProtoData))
		this.Syslog("[SubnetManager.msgParseTCPConn] "+
			"Server 收到测试消息 MsgLen[%d] No.[%d]",
			msgbin.GetTotalLength(), recvmsg.Testno)
		return
	case servercomm.STimeTickCommandID:
		recvmsg := &servercomm.STimeTickCommand{}
		recvmsg.ReadBinary([]byte(msgbin.ProtoData))
		return
	case servercomm.SLoginRetCommandID:
		this.connectMutex.Lock()
		defer this.connectMutex.Unlock()
		// 收到登陆服务器返回的消息
		recvmsg := &servercomm.SLoginRetCommand{}
		recvmsg.ReadBinary([]byte(msgbin.ProtoData))
		if recvmsg.Loginfailed > 0 {
			conn.Terminate()
			if recvmsg.Loginfailed == servercomm.LOGINRETCODE_IDENTICAL {
				conn.IsNormalDisconnect = true
				this.Syslog("[SubnetManager.msgParseTCPConn] "+
					"重复连接,不必连接 TmpID[%s]", conn.GetTempID())
			} else {
				this.Error("[SubnetManager.msgParseTCPConn] "+
					"连接验证失败,断开连接 TmpID[%s]", conn.GetTempID())
			}
			return
		}
		conn.ModuleInfo = recvmsg.Destination
		this.Syslog("[SubnetManager.msgParseTCPConn] "+
			"连接服务器验证成功,id:%s,ipport:%s",
			conn.ModuleInfo.ModuleID, conn.ModuleInfo.ModuleAddr)
		this.subnetHook.OnServerJoinSubnet(conn)
		return
	case servercomm.SLoginCommandID:
		recvmsg := &servercomm.SLoginCommand{}
		recvmsg.ReadBinary([]byte(msgbin.ProtoData))
		this.OnServerLogin(conn, recvmsg)
		return
	case servercomm.SLogoutCommandID:
		// 服务器已主动关闭，不再尝试连接它了
		conn.IsNormalDisconnect = true
		this.connectMutex.Lock()
		defer this.connectMutex.Unlock()
		this.connInfos.Delete(conn.ModuleInfo.ModuleID)
		this.Syslog("[msgParseTCPConn] 服务器已主动关闭，不再尝试连接它了 "+
			"ModuleInfo[%s]", conn.ModuleInfo.GetJson())
		return
	case servercomm.SNotifyAllInfoID:
		// 收到所有服务器的配置信息
		recvmsg := &servercomm.SNotifyAllInfo{}
		recvmsg.ReadBinary([]byte(msgbin.ProtoData))
		this.connectMutex.Lock()
		defer this.connectMutex.Unlock()
		this.Syslog("[SubnetManager.msgParseTCPConn] " +
			"收到所有服务器列表信息")
		// 所有服务器信息列表
		for i := 0; i < len(recvmsg.ServerInfos); i++ {
			serverinfo := recvmsg.ServerInfos[i]
			this.connInfos.Add(serverinfo)
		}
		return
	}
	msgqueues := &ConnectMsgQueueStruct{}
	msgqueues.conn = conn
	msgqueues.msg = msgbin
	this.MultiQueueControl(msgqueues)
}

// 分配消息处理线程
func (this *SubnetManager) MultiQueueControl(
	msgqueues *ConnectMsgQueueStruct) {
	if this.maxRunningMsgNum < 1 {
		this.OnRecvTCPMsg(msgqueues.conn, msgqueues.msg)
		return
	}
	who := this.getRecvTCPMsgParseChan(msgqueues.conn,
		this.maxRunningMsgNum, msgqueues.msg)
	if who >= int32(len(this.runningMsgChan)) || who < 0 {
		panic(fmt.Sprintf("who[%d] >= len(this.runningMsgChan)[%d]", who,
			len(this.runningMsgChan)))
	}
	this.runningMsgChan[who] <- msgqueues
}

// 初始化消息处理队列
func (this *SubnetManager) InitMsgQueue(sum int32) {
	// 最大同时处理的消息数量
	this.maxRunningMsgNum = sum // 消息队列线程数
	if this.maxRunningMsgNum < 1 {
		this.maxRunningMsgNum = 1
		this.Error("[SubnetManager.InitMsgQueue] " +
			"消息处理线程数量过小，置为1...")
	}
	this.runningMsgChan = make([]chan *ConnectMsgQueueStruct,
		this.maxRunningMsgNum)
	this.Syslog("[SubnetManager.InitMsgQueue] "+
		"Task 消息处理线程数量 ThreadNum[%d]", this.maxRunningMsgNum)
	for i := int32(0); i < this.maxRunningMsgNum; i++ {
		this.runningMsgChan[i] = make(chan *ConnectMsgQueueStruct,
			15000)
		go this.RecvmsgProcess(i)
	}
}

// 并行处理接收消息队列数据
func (this *SubnetManager) MultiRecvmsgQueue(
	index int32) (normalreturn bool) {
	if this.runningMsgChan == nil || this.runningMsgChan[index] == nil {
		panic(fmt.Sprintf("this.runningMsgChan[%d] == nil", index))
		return true
	}
	defer func() {
		// 必须要先声明defer，否则不能捕获到panic异常
		if err, stackInfo := sysutil.GetPanicInfo(recover()); err != nil {
			// 这里的err其实就是panic传入的内容
			this.Error("[SubnetManager.MultiRecvmsgQueue] "+
				"Panic: ErrName[%v] \n Stack[%s]", err, stackInfo)
			normalreturn = false
		}
	}()
	normalreturn = true

	msgchan := this.runningMsgChan[index]
	for msgqueues := range msgchan {
		functiontime := monitor.FunctionTime{}
		functiontime.Start("MultiRecvmsgQueue")
		this.OnRecvTCPMsg(msgqueues.conn, msgqueues.msg)
		functiontime.Stop()
	}
	return true
}

// 保持服务器消息处理线程
func (this *SubnetManager) RecvmsgProcess(index int32) {
	for {
		if this.MultiRecvmsgQueue(index) {
			// 正常退出
			break
		}
	}
}
