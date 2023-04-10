# minirpc
一个简单的rpc

之前写了一个rpc框架，由于是第一次写，所以当时想把限流，熔断，服务注册与发现等功能都加进去，当时并没有太关注性能，而且代码很多很乱

所以新写一个minirpc，此项目不会有太多的功能，只有最基础的功能，参考 go-kratos 可以知道，我们完全可以基于一个简单的rpc框架扩展成一个很好用的框架

## 压测
既然此框架注重性能，那么要做好压测，至于怎么做压测，可以看我之前写的文章：[压测](https://xjip3se76o.feishu.cn/wiki/wikcne3GYIP9i952pURS7Vxuhhe)

至于此项目如何做压测，我们直接fork[鸟窝老师的压测仓库](https://github.com/rpcxio/rpcx-benchmark)，模仿他的代码，为自己的rpc框架增加上[压测代码](https://github.com/dayueba/rpc-benchmark)

## 常用优化
我们基于tcp协议，可以很简单的实现一个支持长连接的rpc。

1. 尽可能的使用对象池，减少内存分配，垃圾回收压力
   1. 除了对象，还有[]byte缓冲区
   2. 可以参考 `fasthttp`，此框架就是把对象复用做到极致了，性能才能那么好
2. 使用goroutine pool，有挺多线程的库，复用goroutine
3. 使用bufio，虽然会增加内存的使用，但是大部分情况下会减少系统调用，弊大于利
4. 考虑场景: 长连接 or 短链接 or 连接多路复用。针对场景做特殊优化
   1. 比如说两个微服务之间调用，可能只需要建立少数的长连接就够用了，即便该服务被好几百个服务调用，几千条连接也够用了
   2. 如果是网关等需要建立大量长连接的场景，而且大部分都是空闲连接的话，std net就不合适了，可以使用 `netpoll` 代替std net
5. pprof
   1. go tool pprof --alloc_objects your-program mem.pprof
6. 可以但没必要
   1. 其实还可以将数据压缩后传送，然后再解压，消耗更多的cpu但是传输更快了