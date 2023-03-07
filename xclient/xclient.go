package xclient

import (
	"context"
	"ezrpc"
	"io"
	"log"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *ezrpc.Option
	mu      sync.Mutex
	clients map[string]*ezrpc.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *ezrpc.Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*ezrpc.Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*ezrpc.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvaiable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		// TODO 这里不能:=否则会出现client为nil的情况
		client, err = ezrpc.XDial(rpcAddr, xc.opt)
		if err != nil {
			// fmt.Println("ezrpc.Xdial failed:", err.Error())
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	// fmt.Println("dial client is ", client)
	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		log.Println("xc call.dial err:", err.Error())
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		log.Println("xc Call.d.Get err:", err.Error())
		return err
	}
	// fmt.Println("call select", rpcAddr)
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

// func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
// 	servers, err := xc.d.GetAll()
// 	if err != nil {
// 		log.Println("xc.Broadcast.d.GetAll error:", err.Error())
// 		return err
// 	}
// 	var wg sync.WaitGroup
// 	var mu sync.Mutex
// 	var e error
// 	replyDone := reply == nil
// 	ctx, cancel := context.WithCancel(ctx)
// 	for _, rpcAddr := range servers {
// 		wg.Add(1)
// 		go func(rpcAddr string) {
// 			defer wg.Done()
// 			var clonedReply interface{}
// 			if reply != nil {
// 				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
// 			}
// 			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
// 			mu.Lock()
// 			if err != nil && e == nil {
// 				e = err
// 				cancel()
// 			}
// 			if err == nil && !replyDone {
// 				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
// 				replyDone = true
// 			}
// 			mu.Unlock()
// 		}(rpcAddr)
// 	}
// 	wg.Wait()
// 	return e
// }

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex // protect e and replyDone
	var e error
	replyDone := reply == nil // if reply is nil, don't need to set value
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
