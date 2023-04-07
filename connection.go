package minirpc

import (
	"bufio"
	"context"
	"net"
	"sync"

	"github.com/dayueba/minirpc/protocol"
)

type Connection struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
	// read chan
	readChan chan *protocol.Message
	// write chan
	writeChan chan []byte
	// state

	closed chan struct{}
	sync.Once
}

func (c *Connection) ReadMessage(ctx context.Context) (*protocol.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-c.readChan:
		return msg, nil
	}
}

func (c *Connection) WriteMessage(ctx context.Context, message *protocol.Message) error {
	data, err := message.Encode()
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.writeChan <- data:
		return nil
	}
}

func wrapConn(rawConn net.Conn) *Connection {
	conn := &Connection{
		Conn:      rawConn,
		rd:        bufio.NewReaderSize(rawConn, 4096), //创建一个大小为4KB的读缓冲
		wr:        bufio.NewWriterSize(rawConn, 1024), //创建一个大小为1KB的写缓冲
		readChan:  make(chan *protocol.Message, 10),
		writeChan: make(chan []byte, 10),
		closed:    make(chan struct{}, 1),
	}

	go conn.readloop()
	go conn.writeloop()
	return conn
}

func (c *Connection) Close() {
	c.Once.Do(func() {
		c.closed <- struct{}{}
		c.Conn.Close()
	})
}

func (c *Connection) readloop() (err error) {
	// todo: writeloop同理，需要防止启动多个readloop/writeloop
	//log := logrus.WithFields(logrus.Fields{
	//	"module": "connection",
	//	"func":   "readloop",
	//})
	//defer func() {
	//	if err != nil && err != io.EOF {
	//		log.Error("connection readloop exited with error: ", err)
	//	} else {
	//		log.Info("connection readloop exited")
	//	}
	//
	//	c.Close()
	//}()

	for {
		select {
		default:
			msg := protocol.NewMessage()
			err = msg.Decode(c.rd)
			if err != nil {
				return
			}
			c.readChan <- msg
		case <-c.closed:
			return
		}
	}
	return
}

func (c *Connection) writeloop() (err error) {
	//log := logrus.WithFields(logrus.Fields{
	//	"module": "connection",
	//	"func":   "writeloop",
	//})
	//defer func() {
	//	if err != nil {
	//		log.Error("connection writeloop exited with error: ", err)
	//	} else {
	//		log.Info("connection writeloop exited")
	//	}
	//	c.Close()
	//}()

	for {
		select {
		case payload := <-c.writeChan:
			_, err = c.wr.Write(payload)
			if err != nil {
				return
			}
			chanlen := len(c.writeChan)
			for i := 0; i < chanlen; i++ {
				payload = <-c.writeChan
				_, err = c.wr.Write(payload)
				if err != nil {
					return
				}
			}
			// 主动Flush数据
			err = c.wr.Flush()
			if err != nil {
				return
			}
		case <-c.closed:
			return
		}
	}
	return
}
