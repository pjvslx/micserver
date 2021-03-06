MICServer
--------------

micserver是一个为分布式系统设计的服务器框架，以模块为服务的基本单位。底层模块间使用TCP及自定义二进制编码格式通信，底层实现优先考虑时间效率。

使用ROC(Remote Object Call)远程对象调用作为模块间的主要通信接口，不关心服务模块本身，而是将上层业务作为对象注册到micserver中，寻址路由均由底层维护，只需要知道目标调用对象的ID即可调用，不需要关心目标所在的模块。得益于ROC抽象，使所有业务状态都可以与服务本身解耦，可以轻松将各个ROC对象在模块间转移或加载，实现分布式系统的**无状态**/**热更**/**容灾冗余**等特性。

你可以在[示例程序](https://github.com/liasece/micchaos)中了解micserver的基本使用方法。

目前micserver不需要任何第三方包。

安装
--------------
    go get github.com/liasece/micserver

官方文档
--------------

[GoDoc](https://godoc.org/github.com/liasece/micserver)
