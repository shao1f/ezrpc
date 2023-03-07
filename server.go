package ezrpc

import (
	"encoding/json"
	"ezrpc/codec"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	MagicNumber = 0x2e3af5
)

type Option struct {
	CodecType      codec.CodecType
	MagicNumber    int
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = Option{
	CodecType:      codec.GobType,
	MagicNumber:    MagicNumber,
	ConnectTimeout: time.Second * 10,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("ezrpc:server listener accept err:", err.Error())
			return
		}
		go server.serverConn(conn)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func (server *Server) serverConn(conn io.ReadWriteCloser) {
	defer conn.Close()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("ezrpc:serverConn json decoder option failed,err:", err.Error())
		return
	}
	// log.Println("opt", opt)
	if opt.MagicNumber != MagicNumber {
		log.Println("ezrpc:serverConn opt.MagicNum is not equal")
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Println("ezrpc:serverConn opt.codecType not supporsed")
		return
	}
	server.serverCodec(f(conn), &opt)
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return fmt.Errorf("ezrpc: service already defined:%s", s.name)
	}
	return nil
}

func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) findService(serviceMethod string) (svc *service, mType *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = fmt.Errorf("ezrpc server: service/method request err:%v", serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = fmt.Errorf("ezrpc server:cant`t find service %v", serviceName)
		return
	}
	svc = svci.(*service)
	mType = svc.method[methodName]
	if mType == nil {
		err = fmt.Errorf("ezrpc server: can`t find method %v", methodName)
	}
	return
}

type Request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	mType        *methodType
	svc          *service
}

var invalidRequest = struct{}{}

func (server *Server) serverCodec(cc codec.Codec, opt *Option) {
	var sending = new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout)
	}
	wg.Wait()
	_ = cc.Close()
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("ezrpc:readRequestHeader err:", err.Error())
		}
		return nil, err
	}
	// log.Println("ezrpc:readRequestHeader:", h)
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*Request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		// log.Println("ezrpc server readRequest err,", err.Error())
		return nil, err
	}
	req := &Request{
		h: h,
	}
	req.svc, req.mType, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mType.newArgv()
	req.replyv = req.mType.newReplyV()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Pointer {
		argvi = req.argv.Addr().Interface()
	}
	if err := cc.ReadBody(argvi); err != nil {
		log.Println("ezrpc,readRequest.cc.ReadBody err:", err.Error())
		return req, err
	}
	return req, nil
}

func (server *Server) handleRequest(cc codec.Codec, req *Request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	// log.Println(req.h, req.argv.Elem())
	called := make(chan struct{})
	sent := make(chan struct{})
	finish := make(chan struct{})
	defer close(finish)
	go func() {
		err := req.svc.call(req.mType, req.argv, req.replyv)
		// called <- struct{}{}
		// if err != nil {
		// 	req.h.Error = err.Error()
		// 	server.sendResponse(cc, req.h, invalidRequest, sending)
		// 	sent <- struct{}{}
		// 	return
		// }
		// server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		// sent <- struct{}{}
		select {
		case <-finish:
			close(called)
			close(sent)
			return
		case called <- struct{}{}:
			if err != nil {
				req.h.Error = err.Error()
				server.sendResponse(cc, req.h, invalidRequest, sending)
				sent <- struct{}{}
				return
			}
			server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
			sent <- struct{}{}
		}
	}()
	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("ezrpc server: request handle timeout:expect within:%s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, b interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, b); err != nil {
		log.Println("ezrpc:sendResponse err:", err.Error())
	}
}

const (
	connected        = "200 Connected to EzRPC"
	defaultRPCPath   = "/_ezrpc_"
	defaultDebugPath = "/debug/ezrpc"
)

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodConnect {
		w.Header().Set("content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Println("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.serverConn(conn)
}

func (server *Server) HandleHTTP() {
	http.Handle(defaultRPCPath, server)
	http.Handle(defaultDebugPath, debugHttp{server})
	log.Println("rpc server debug path:", defaultDebugPath)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
