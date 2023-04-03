# minirpc
一个简单的rpc

之前写了一个rpc框架，由于是第一次写，所以当时想把限流，熔断，服务注册与发现等功能都加进去，当时并没有太关注性能，而且代码很多很乱

所以新写一个minirpc，此项目不会有太多的功能，只有最基础的功能，参考 go-kratos 可以知道，我们完全可以基于一个简单的rpc框架扩展成一个很好用的框架

## 压测
既然此框架注重性能，那么要做好压测，至于怎么做压测，可以看我之前写的文章：[压测](https://xjip3se76o.feishu.cn/wiki/wikcne3GYIP9i952pURS7Vxuhhe)

至于此项目如何做压测，我们直接fork[鸟窝老师的压测仓库](https://github.com/rpcxio/rpcx-benchmark)，模仿他的代码，为自己的rpc框架增加上[压测代码](https://github.com/dayueba/rpc-benchmark)

## 常用优化
我们基于tcp协议，可以很简单的实现一个支持长连接的rpc。

1. 使用sync.Pool 优化
   - Call