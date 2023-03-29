package minirpc

import (
	"errors"
	"io"
	"log"
	"minirpc/protocol"
	"net"
	"sync"
)

// ServerError represents an error that has been returned from
// the remote side of the RPC connection.
type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")

type Call struct {
	ServicePath   string
	ServiceMethod string
	Args          any
	Reply         any
	Error         error
	Done          chan *Call
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
	}
}

type Client interface {
	// Call 同步调用
	Call(path, method string, args any, reply any) error
	// Go 异步调用
	Go(path, method string, args any, reply any, done chan *Call) *Call
}

var _ Client = (*client)(nil)

type client struct {
	conn net.Conn

	reqMutex sync.Mutex
	mutex    sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

func NewClient(addr string) (*client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &client{conn: conn, pending: make(map[uint64]*Call)}
	go c.readloop()
	return c, nil
}

func (c *client) readloop() {
	var err error
	for err == nil {
		res := protocol.NewMessage()
		err := res.Decode(c.conn)
		if err != nil {
			break
		}
		seq := res.Seq()
		c.mutex.Lock()
		call := c.pending[seq]
		delete(c.pending, seq)
		c.mutex.Unlock()

		switch {
		case call == nil:
			// 没有正在等待响应的请求，出现这个问题的原因一般是因为 WriteRequest 部分失败，导致我们删除了这个call
			// 虽然我们不需要处理这个请求，但是需要把数据读取了
			// 因为前面已经读取了，所以这里什么都不需要做
		case res.MessageStatusType() == protocol.Error:
			// todo：rpc返回错误目前先忽略
		default:
			data := res.Payload
			if len(data) > 0 {
				err = msgpackCodec.Decode(data, call.Reply)
				if err != nil {
					call.Error = err
				}
			}
			call.done()
		}
	}
	// Terminate pending calls.
	c.reqMutex.Lock()
	c.mutex.Lock()
	c.shutdown = true
	closing := c.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
	c.mutex.Unlock()
	c.reqMutex.Unlock()
	if err != io.EOF && !closing {
		log.Println("rpc: client protocol error:", err)
	}
}

func (c *client) Call(servicePath, serviceMethod string, args any, reply any) error {
	call := <-c.Go(servicePath, serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}

func (c *client) Go(servicePath, serviceMethod string, args any, reply any, done chan *Call) *Call {
	call := &Call{
		ServiceMethod: serviceMethod,
		ServicePath:   servicePath,
		Args:          args,
		Reply:         reply,
	}
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// 传入的done必须有缓冲区，否则最好不要运行
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	c.send(call)
	return call
}

func (c *client) Close() error {
	c.mutex.Lock()
	if c.closing {
		c.mutex.Unlock()
		return ErrShutdown
	}
	c.closing = true
	c.mutex.Unlock()
	//return client.codec.Close()
	return nil
}

func (c *client) send(call *Call) {
	c.reqMutex.Lock()
	defer c.reqMutex.Unlock()

	// 缓存/注册 此次请求
	c.mutex.Lock()
	if c.shutdown || c.closing {
		c.mutex.Unlock()
		call.Error = ErrShutdown
		call.done()
		return
	}

	seq := c.seq
	c.seq++
	c.pending[seq] = call
	c.mutex.Unlock()

	// 发送请求
	//c.request.Seq = seq
	//c.request.ServiceMethod = call.ServiceMethod
	//err := c.codec.WriteRequest()
	req := protocol.NewMessage()
	req.SetSeq(seq)
	req.ServicePath = call.ServicePath
	req.ServiceMethod = call.ServiceMethod

	data, err := msgpackCodec.Encode(call.Args)
	if err != nil {
		c.mutex.Lock()
		delete(c.pending, seq)
		c.mutex.Unlock()
		call.Error = err
		call.done()
		return
	}
	req.Payload = data

	allData, err := req.Encode()
	if err != nil {
		c.mutex.Lock()
		delete(c.pending, seq)
		c.mutex.Unlock()
		call.Error = err
		call.done()
		return
	}

	_, err = c.conn.Write(allData)
	if err != nil {
		c.mutex.Lock()
		call = c.pending[seq]
		delete(c.pending, seq)
		c.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}
}
