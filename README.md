## mdgo


mdgo是一个基于事件循环的网络库，遵循reactor模型。是`mainReactor+subReactor`模式。`mainReactor`负责检测监听`socket`。当监听`socket`可读，则将已连接套接字，投递给`subReactor`。因为遵循`一个线程(go协程)一个事件循环`，因而避免了共享资源竞争，将锁的开销降低到最低。

![架构图](./doc/mdgo.png)
