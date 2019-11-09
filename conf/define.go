package conf

type ConfigKey string

var (
	// 版本号		string
	Version ConfigKey = "version"
	// 进程ID		string
	ProcessID ConfigKey = "processid"
	// log完整路径		string
	LogWholePath ConfigKey = "logpath"
	// 服务器TCP子网ip及端口，如 1.0.0.1:80 		string
	SubnetTCPAddr ConfigKey = "subnettcpaddr"
	// 不使用本地chan		bool
	SubnetNoChan ConfigKey = "subnetnochan"
	// 网关TCP地址		string
	GateTCPAddr ConfigKey = "gatetcpaddr"
	// 是否是受保护的进程 		bool
	IsDaemon ConfigKey = "isdaemon"
	// 消息处理并发协程数量		int
	MsgThreadNum ConfigKey = "msgthreadnum"
)