## mdgo

mdgo是一个基于事件循环的网络库，遵循reactor模型。是`mainReactor+subReactor`模式。`mainReactor`负责检测监听`socket`。当监听`socket`可读，则将已连接套接字，投递给`subReactor`。因为遵循`一个线程(go协程)一个事件循环`，因而避免了共享资源竞争，将锁的开销降低到最低。

![架构图](./doc/mdgo.png)


### server.go

~~~go
type Server struct {
	started       atomic.Bool    // 标明服务器是否启动
	option        *Option        // 配置选项
	handler       Handler        // 回调句柄
	codec         Codec          // 编解码器 ??? 是否可以放到 EventLoop 减少锁开销
	mainLoop      *EventLoop     // mainReactor
	workLoops     []*EventLoop   // subReactor
	nextLoopIndex int            // workLoop索引
	wg            sync.WaitGroup // 同步
}
~~~
这个结构体是`mdgo`的核心。包含3个核心函数：

1. 新建服务器
2. 服务器启动
3. 服务器停止

2个私有函数：
1. 获取下一个`eventLoop`
2. 当新的`acceptFd`到来的回调处理

### listener.go / connection.go

在`mdgo`中，`listener`和`connection`是平级的关系。

1. listener : 代表监听套接字
2. connection : 代表已连接套接字，是对TCP连接的抽象

~~~go
type Listener struct {
	listenFd      int             // 监听套接字
	file          *os.File        // dupFd
	handleNewConn HandlerConnFunc // new acceptFd coming to handle
	listener      net.Listener    // net.Listener
	loop          *EventLoop      // pointer main eventloop
}
~~~

`Listener`是监听套接字的处理模块，对于监听套接字，事件循环只处理`可读事件`,当监听套接字可读，即表示可以使用`accept`来获取已连接套接字，也表示`TCP三次握手完成`。之后将`acceptFd`嵌入`Connection`。

~~~go
type Connection struct {
	connFd     int               // acceptFd
	connected  atomic.Bool       // state[connected or not]
	InBuf      *buffer.FixBuffer // input buffer
	OutBuf     *buffer.FixBuffer // output buffer
	cb         Callback          // cb
	peerAddr   string            // remote addr
	eventLoop  *EventLoop        // work sub eventLoop
	activeTime atomic.Int64      // last active time
}
~~~

`Connection`是已连接套接字处理模块，是对`tcp三次握手`的抽象，当三次握手建立完成后，通过`Connection`进行数据的接收和发送，必要的时候对`Connection`进行关闭操作。

综上，通过`Listener`和`Connection`, 处理连接的三个半事件。

### eventLoop.go

`eventLoop`是事件循环的核心，负责新建事件循环，事件循环启动，等待套接字事件就绪。  

### eventHolder

`mdgo`封装了自己的`event`包裹器,命名为`eventHolder`, 包含`监听的fd`,`关注的事件`,`已经就绪的事件`。这样就能很好的封装之后的轮询器`poller`。不关系底层，统一暴露`eventHolder`,供事件循环器`eventloop`使用。

### poller

`poller`是轮询器文件夹，根据不同的系统，分装了不同的系统函数，例如`linux`下的`epoll`, `MacOs`下的`kqueue`等。


