package manager

import (
	"github.com/liasece/micserver/log"
	"github.com/liasece/micserver/server/subnet"
	"github.com/liasece/micserver/tcpconn"
	"net"
	"time"
)

// websocket连接管理器
type ClientSocketManager struct {
	connPool      tcpconn.ClientConnPool
	subnetManager *subnet.SubnetManager

	Analysiswsmsgcount uint32
}

func (this *ClientSocketManager) Init(subnetManager *subnet.SubnetManager, n int) {
	this.subnetManager = subnetManager
	this.connPool.Init(n)
}

func (this *ClientSocketManager) AddClientTcpSocket(
	conn net.Conn) (*tcpconn.ClientConn, error) {
	task, err := this.connPool.NewClientConn(conn)
	if err != nil {
		return nil, err
	}
	curtime := uint64(time.Now().Unix())
	task.SetTerminateTime(curtime + 20) // 20秒以后还没有验证通过就断开连接

	task.Debug("[ClientSocketManager.AddClientTcpSocket] "+
		"新增连接数 当前连接数量 NowSum[%d]",
		this.GetClientTcpSocketCount())
	return task, nil
}

// 根据 OpenID 索引 Task
func (this *ClientSocketManager) AddTaskOpenID(
	task *tcpconn.ClientConn, openid string) {
	this.connPool.AddTaskOpenID(task, openid)
}

// 根据 OpenID 索引 Task
func (this *ClientSocketManager) GetTaskByOpenID(
	openid string) *tcpconn.ClientConn {
	return this.connPool.GetTaskByOpenID(openid)
}

// 根据 UUID 索引 Task
func (this *ClientSocketManager) AddTaskUUID(task *tcpconn.ClientConn,
	uuid uint64) {
	this.connPool.AddTaskUUID(task, uuid)
}

// 根据 UUID 索引 Task
func (this *ClientSocketManager) GetTaskByUUID(
	uuid uint64) *tcpconn.ClientConn {
	return this.connPool.GetTaskByUUID(uuid)
}

func (this *ClientSocketManager) GetTaskByTmpID(
	webtaskid uint64) *tcpconn.ClientConn {
	return this.connPool.Get(webtaskid)
}

func (this *ClientSocketManager) GetClientTcpSocketCount() uint32 {
	return this.connPool.Len()
}

func (this *ClientSocketManager) remove(tempid uint64) {
	value := this.GetTaskByTmpID(tempid)
	if value == nil {
		return
	}
	// 设置redis缓存
	if value.Openid != "" && value.Userserverid != 0 {
		subnet.GetGBRedisManager().
			AddUserServerIDByOpenidWithDeadline(value.Openid,
				value.Userserverid, 1*60*60)
	}
	this.connPool.Remove(tempid)
}

func (this *ClientSocketManager) RemoveTaskByTmpID(
	tempid uint64) {
	this.remove(tempid)
}

// 遍历所有的连接
func (this *ClientSocketManager) ExecAllUsers(
	callback func(uint64, *tcpconn.ClientConn)) {
	this.connPool.Range(func(value *tcpconn.ClientConn) {
		callback(value.Tempid, value)
	})
}

// 遍历所有的连接，检查需要移除的连接
func (this *ClientSocketManager) ExecRemove(
	callback func(*tcpconn.ClientConn) bool) {
	removelist := make([]uint64, 0)
	this.connPool.Range(func(value *tcpconn.ClientConn) {
		// 遍历所有的连接
		if callback(value) {
			// 该连接需要被移除
			removelist = append(removelist, value.Tempid)
			value.Terminate()
		}
	})
	for _, v := range removelist {
		this.remove(v)
	}

	log.Debug("[ClientSocketManager.ExecRemove] "+
		"条件删除连接数 RemoveSum[%d] 当前连接数量 NowSum[%d]",
		len(removelist), this.GetClientTcpSocketCount())
}

// func NotifyClientUserOffline(task *tcpconn.ClientConn) {
// 	if task == nil || len(task.Openid) == 0 || !task.IsVertify() {
// 		return
// 	}
// 	task.Debug("[Socket] 用户下线,断开连接了")
// 	sendmsg := &comm.SGatewayWSOfflineUser{}
// 	sendmsg.Openid = task.Openid
// 	sendmsg.UUID = task.UUID
// 	sendmsg.ClientConnID = task.Tempid
// 	userclient := this.subnetManager.GetClientManager().
// 		GetTCPClient(uint64(task.Userserverid))
// 	if userclient != nil {
// 		userclient.SendCmd(sendmsg)
// 	}
// 	roomclient := this.subnetManager.GetClientManager().
// 		GetTCPClient(uint64(task.Roomserverid))
// 	if roomclient != nil {
// 		roomclient.SendCmd(sendmsg)
// 	}
// 	subnet.GetSubnetManager().GetServerManger().NotifyOtherMyInfo()
// }