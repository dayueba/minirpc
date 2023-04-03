package minirpc

import (
	"context"
	"errors"
	"go/token"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dayueba/minirpc/protocol"
)

type Server struct {
	opts       *ServerOptions
	serviceMap sync.Map   // map[string]*service
	reqLock    sync.Mutex // protects freeReq
	respLock   sync.Mutex // protects freeResp
}

func NewServer(addr string, opts ...ServerOption) *Server {
	s := &Server{
		opts: &ServerOptions{addr: addr},
	}

	for _, o := range opts {
		o(s.opts)
	}

	return s
}

func (server *Server) Register(rcvr any) error {
	return server.register(rcvr, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr any) error {
	return server.register(rcvr, name, true)
}

// logRegisterError specifies whether to log problems during method registration.
// To debug registration, recompile the package with this set to true.
const logRegisterError = false

func (server *Server) register(rcvr any, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Print(s)
		return errors.New(s)
	}
	if !token.IsExported(sname) && !useName {
		s := "rpc.Register: type " + sname + " is not exported"
		log.Print(s)
		return errors.New(s)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, logRegisterError)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PointerTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.Print(str)
		return errors.New(str)
	}

	if _, dup := server.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

func (server *Server) Start() error {
	lis, err := net.Listen("tcp", server.opts.addr)
	if err != nil {
		return err
	}

	var tempDelay time.Duration

	tl, ok := lis.(*net.TCPListener)
	if !ok {
		return errors.New("NetworkNotSupportedError")
	}

	for {
		conn, err := tl.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		if err = conn.SetKeepAlive(true); err != nil {
			return err
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					logrus.WithFields(logrus.Fields{
						"func": "handleConn",
					}).Errorf("handle tcp conn panic error, %s", err)
				}
			}()

			if err := server.handleConn(conn); err != nil {
				logrus.WithFields(logrus.Fields{
					"func": "handleConn",
				}).Errorf("handle tcp conn error, %s", err)
			}
		}()
	}
}

func (server *Server) handleConn(conn net.Conn) error {
	defer conn.Close()

	for {
		ctx := context.Background()
		if server.opts.timeout > 0 {
			// todo: timeout
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), server.opts.timeout)
			defer cancel()
		}

		req := protocol.NewMessage()
		err := req.Decode(conn)
		if err == io.EOF {
			// read compeleted
			return nil
		}

		if err != nil {
			return err
		}

		//go func() {
		//
		//	//rsp, err := server.handle(ctx, frame)
		//	//if err != nil {
		//	//	logrus.WithFields(logrus.Fields{
		//	//		"func": "handleConn",
		//	//	}).Errorf("s.handle err is not nil, %s", err)
		//	//}
		//	//
		//	//if err = server.write(ctx, conn, rsp); err != nil {
		//	//	logrus.WithFields(logrus.Fields{
		//	//		"func": "handleConn",
		//	//	}).Errorf("server write res err, %s", err)
		//	//}
		//	go server.handleRequest(ctx, req, conn)
		//}()

		go server.handleRequest(ctx, req, conn)
	}
}

func (server *Server) handleRequest(ctx context.Context, req *protocol.Message, conn net.Conn) {
	log := logrus.WithFields(logrus.Fields{
		"func": "handleRequest",
	})
	defer func() {
		if err := recover(); err != nil {
			// todo: deal error
		}
	}()
	var err error
	res := req.Clone()
	// Look up the request.
	svci, ok := server.serviceMap.Load(req.ServicePath)
	if !ok {
		err = errors.New("rpc: can't find service " + req.ServiceMethod)
		return
	}
	svc := svci.(*service)
	mtype := svc.method[req.ServiceMethod]
	if mtype == nil {
		err = errors.New("rpc: can't find method " + req.ServiceMethod)
	}
	if err != nil {
		// todo: deal error
		log.Error(err)
	}
	argv := reflectTypePools.Get(mtype.ArgType)
	codec, ok := Codecs[req.SerializeType()]
	if !ok {
		// todo: 序列化类型错误
		log.Error("没有找到codec", req.SerializeType())
	}
	err = codec.Decode(req.Payload, argv)
	if err != nil {

	}
	replyv := reflectTypePools.Get(mtype.ReplyType)

	if mtype.ArgType.Kind() != reflect.Ptr {
		err = svc.call(ctx, mtype, reflect.ValueOf(argv).Elem(), reflect.ValueOf(replyv))
	} else {
		err = svc.call(ctx, mtype, reflect.ValueOf(argv), reflect.ValueOf(replyv))
	}
	reflectTypePools.Put(mtype.ArgType, argv)
	res.Payload, err = codec.Encode(replyv)
	defer reflectTypePools.Put(mtype.ReplyType, replyv)
	if err != nil {

	}
	allData, err := res.Encode()
	if err != nil {

	}

	conn.Write(allData)
}

// todo: handle error
func (server *Server) handleError(res *protocol.Message, err error) (*protocol.Message, error) {
	res.SetMessageStatusType(protocol.Error)
	return res, err
}

func (server *Server) sendResponse(ctx context.Context, conn net.Conn, err error, req, res *protocol.Message) {
	data, _ := res.Encode()
	conn.Write(data)
}
