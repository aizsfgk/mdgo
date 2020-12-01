package callback


//// 回调函数
//type MessageCallback func(int64)
//type WriteCompleteCallback func(int64)
//type CloseCallback func(int64)


//
// GO 语言中使用特殊的Interface 实现特殊逻辑
//
type Callback interface {
	OnMessage(int64)
	OnClose()
	OnWriteComplete()
}
