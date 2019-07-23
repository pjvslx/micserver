package comm

type SServerInfo struct {
	ServerID   string
	ServerAddr string
	// 服务器序号 重复不影响正常运行
	// 但是其改动会影响 配置读取/ServerName/Log文件名
	ServerNumber uint32
	// 服务器数字版本
	// 命名规则为： YYYYMMDDhhmm (年月日时分)
	Version uint64
}

type STimeTickCommand struct {
	Testno uint32
}

type STestCommand struct {
	Testno     uint32
	Testttring string // IP
}

type SLoginCommand struct {
	ServerID   string
	ServerAddr string // IP
	// 登录优先级
	ConnectPriority int64
	// 服务器序号 重复不影响正常运行
	// 但是其改动会影响 配置读取/ServerName/Log文件名
	ServerNumber uint32
	// 服务器数字版本
	// 命名规则为： YYYYMMDDhhmm (年月日时分)
	Version uint64
}

// 通知服务器正常退出
type SLogoutCommand struct {
}

// 通知我所连接的服务器启动成功
type SSeverStartOKCommand struct {
	Serverid uint32
}

const (
	// 登录成功
	LOGINRETCODE_SECESS = 0
	// 身份验证错误
	LOGINRETCODE_IDENTITY = 1
	// 重复连接
	LOGINRETCODE_IDENTICAL = 2
)

// 登录服务器返回
type SLoginRetCommand struct {
	Loginfailed uint32      // 是否连接成功,0成功
	Destination SServerInfo //	tcptask 所在服务器信息
}

// super通知其它服务器启动成功
type SStartRelyNotifyCommand struct {
	Serverinfos []SServerInfo // 启动成功服务器信息
}

// 启动验证通过通知其它服务器我的新
type SStartMyNotifyCommand struct {
	Serverinfo SServerInfo // 启动成功服务器信息
}

// super 通知所有服务器配置信息
type SNotifyAllInfo struct {
	Serverinfos []SServerInfo // 成功服务器信息
}
type SUpdateGatewayUserAnalysis struct {
	Httpcount         uint32 // 5分钟的量
	Webscoketcount    uint32 // 5分钟的量
	Webscoketcurcount uint32 // 当前的量
}

// 通知super 新添加用户 存入redis
type SAddNewUserToRedisCommand struct {
	Openid       string
	Serverid     uint32
	UUID         uint64
	ClientConnID uint64
}

// gateway 转发过来的消息数据 gateway <<-->> QuizServer
type SGatewayForwardCommand struct {
	Gateserverid uint32
	ClientConnID uint64
	Openid       string
	UUID         uint64
	Cmdid        uint16
	Cmdlen       uint16
	Cmddatas     []byte
}

// gateway 广播转发消息数据 gateway <<-->> clients
type SGatewayForwardBroadcastCommand struct {
	Gateserverid uint32
	// 处理线程哈希，需要提供给Gateway作发送协程选择，
	// 合理配置该项可以保证消息达到顺序且提高gateway性能，否则，
	// 两条广播消息的到达顺序无法保证！！！
	ThreadHash uint32
	UUIDList   []uint64
	Cmdid      uint16
	Cmdlen     uint16
	Cmddatas   []byte
}

// 向Gateway发送GM返回数据，Userserver -> Gatewayserver
type SGatewayForward2HttpCommand struct {
	Gateserverid uint32
	Httptaskid   uint64
	Openid       string
	Cmdname      string
	Cmdlen       uint16
	Cmddatas     []byte
}

//  转发消息给指定的userserver  user->bridge->user 直接给用户
type SBridgeForward2UserCommand struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	Cmdlen   uint16
	Cmddatas []byte
}

//  广播消息给所有的userserver  userserver->bridge->user 直接给用户
type SBridgeBroadcast2UserCommand struct {
	Fromopenid string
	Toopenid   string
	Cmdid      uint16
	Cmdlen     uint16
	Cmddatas   []byte
}

//  转发消息给指定的userserver  userserver->bridge->userserver 直接给用户所在的server
type SBridgeForward2UserServer struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	Cmdlen   uint16
	Cmddatas []byte
}

//  转发消息给所有 GatewayServer
//  GatewayServer->bridge->GatewayServer 直接给用户Client
type SBridgeBroadcast2GatewayServer struct {
	Cmdid    uint16
	Cmdlen   uint16
	Cmddatas []byte
}

//  发送消息给指定的userserver  matchserver->userserver 直接给用户所在的server
type SMatchForward2UserServer struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	Cmdlen   uint16
	Cmddatas []byte
}

//  发送消息给指定的userserver  roomserver->userserver 直接给用户所在的server
type SRoomForward2UserServer struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	Cmdlen   uint16
	Cmddatas []byte
}

// 网关广播消息发送给每个玩家
type SGatewayBroadcast2UserCommand struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	Cmdlen   uint16
	Cmddatas []byte
}

// 请求查询指定玩家的好友公开信息 方便搜索好友
type SUserServerSearchFriend struct {
	Fromuuid uint64
	Touuid   uint64
}

// 执行gm指令
type SUserServerGMCommand struct {
	Taskid uint64
	Key    string
	UUID   uint64
	Openid string
	CmdID  uint32
	Param1 string
	Param2 string
}

// 请求远程执行另外一个玩家的操作相关
type SRequestOtherUser struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	CmdData  []byte
}

// 响应另一个玩家的操作
type SResponseOtherUser struct {
	Fromuuid uint64
	Touuid   uint64
	Cmdid    uint16
	CmdData  []byte
}

// 去bridge获得一个随机的玩家信息 返回User
type SBridgeDialGetUserInfo struct {
	Fromopenid string
	Fromuuid   uint64
	Getuuid    uint64
	Getopenid  string
	Type       uint32
}

// 这个消息之前必定已经经过http login user
// gateway 发起用户登录 Gateway-->UserServer ->QuizServer-->Gateway
type SGatewayWSLoginUser struct {
	Gateserverid uint32
	ClientConnID uint64
	Openid       string
	UUID         uint64
	Token        string
	Tokenendtime uint32
	Sessionkey   string
	Loginappid   string
	Username     string
	Quizid       uint64
	Allmoney     uint64
	Headurl      string
	Female       uint32 //是否女性玩家  1 女玩家 0男玩家
	Retcode      int32  // 0 是成功，非0都表示失败了
	Message      string
	LoginMsg     []byte // 原始登陆消息
	// retcode 1 服务器异常
	// retcode 2 用户登录异常
}

// gateway 发起用户下线 Gateway-->UserServer ->QuizServer
type SGatewayWSOfflineUser struct {
	Openid       string
	UUID         uint64
	Quizid       uint64
	ClientConnID uint64
}

type STemplateMessageKeyWord struct {
	Value string
	Color string
}

// QuizServer通知给玩家发送模板消息
type SQSTemplateMessage struct {
	Openid      string
	Template_id string
	Page        string
	Datalist    []STemplateMessageKeyWord
	Formid      string
}

// 网关通过super广播换accesstoken
type SGatewayChangeAccessToken struct {
	Access_token         string //  微信access_token
	Update_accesstime    uint32 // 需要更新access_token 的时间
	Access_token_QQ      string //  QQaccess_token
	Update_accesstime_QQ uint32 // 需要更新QQ access_token 的时间
}

//  广播消息给所有的userserver  userserver->matchserver->userserver
type SMatchBroadcast2UserServerCommand struct {
	Fromuuid   uint64
	Matchindex uint64 // 匹配队列索引，即RoomID
	Cmdid      uint16
	Cmdlen     uint16
	Cmddatas   []byte
}

type SRequestServerInfo struct {
}

// 由 SuperServer 发送给其他服务器
// 通知说明的目标服务器安全退出
// 此消息发送的前提是当前存在可以替代目标服务器的服务器
type SNotifySafelyQuit struct {
	// 目标服务器的信息应该是最新的信息，目标服务器会将该信息改成最新的
	TargetServerInfo SServerInfo
}
