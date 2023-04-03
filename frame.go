package minirpc

import (
	"context"
	"net"

	"github.com/dayueba/minirpc/protocol"
)

const defaultReadChanSize = 10
const defaultWriteChanSize = 10

type Connection struct {
	net.Conn
	// read chan
	readChan chan *protocol.Message
	// write chan
	writeChan chan *protocol.Message
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.writeChan <- message:
		return nil
	}
}

func wrapConn(rawConn net.Conn) *Connection {
	conn := &Connection{
		Conn:      rawConn,
		readChan:  make(chan *protocol.Message, defaultReadChanSize),
		writeChan: make(chan *protocol.Message, defaultWriteChanSize),
	}

	go conn.readloop()
	go conn.writeloop()
	return conn
}

func (c *Connection) readloop() error {
	//for {
	//frameHeader := make([]byte, FrameHeadLen)
	//if num, err := io.ReadFull(c.Conn, frameHeader); num != FrameHeadLen || err != nil {
	//	return err
	//}
	//
	//length := binary.BigEndian.Uint32(frameHeader) // 目前header里只有length
	//
	////if length > MaxPayloadLength {
	////	return nil, codes.NewFrameworkError(codes.ClientMsgErrorCode, "payload too large...")
	////}
	//
	//for uint32(len(f.buffer)) < length && f.counter <= 12 {
	//	f.buffer = make([]byte, len(f.buffer)*2)
	//	f.counter++
	//}
	//
	//if num, err := io.ReadFull(conn, f.buffer[:length]); uint32(num) != length || err != nil {
	//	return nil, err
	//}
	//
	//return append(frameHeader, f.buffer[:length]...), nil
	//}
	return nil
}

func (c *Connection) writeloop() error {
	return nil
}
