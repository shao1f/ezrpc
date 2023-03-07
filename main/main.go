package main

import (
	"context"
	"ezrpc"
	"ezrpc/registry"
	"ezrpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// func startServer(addr chan string) {
// 	l, err := net.Listen("tcp", ":8779")
// 	if err != nil {
// 		log.Panic("net.Listen err:", err.Error())
// 	}
// 	log.Println("start ezrpc server on:", l.Addr())
// 	addr <- l.Addr().String()
// 	ezrpc.Accept(l)
// }

// func main() {
// 	addr := make(chan string)
// 	go startServer(addr)

// 	conn, _ := net.Dial("tcp", <-addr)
// 	defer conn.Close()

// 	time.Sleep(time.Second)
// 	_ = json.NewEncoder(conn).Encode(ezrpc.DefaultOption)
// 	cc := codec.NewGobCodec(conn)
// 	for i := 0; i < 5; i++ {
// 		h := &codec.Header{
// 			ServiceMethod: "EzRpc.Main",
// 			Seq:           uint64(i),
// 		}
// 		err := cc.Write(h, fmt.Sprintf("ezrpc req:%v", h.Seq))
// 		if err != nil {
// 			log.Println("cc.Writer err:", err.Error())
// 		}
// 		err = cc.ReadHeader(h)
// 		if err != nil {
// 			log.Println("cc.ReadHeader err", err.Error())
// 		}
// 		// fmt.Println(h)
// 		var reply string
// 		err = cc.ReadBody(&reply)
// 		if err != nil {
// 			log.Println("cc.ReadBody err:", err.Error())
// 		}
// 		log.Println("reply:", reply)
// 	}

// }

// func main() {
// 	log.SetFlags(log.Lshortfile | log.LstdFlags)
// 	addr := make(chan string)
// 	go startServer(addr)

// 	client, err := ezrpc.Dail("tcp", <-addr)
// 	if err != nil {
// 		fmt.Println("client dial err:", err.Error())
// 	}
// 	defer func() {
// 		_ = client.Close()
// 	}()

// 	time.Sleep(time.Second * 1)

// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			args := fmt.Sprintf("ezrpc req %d", i)
// 			var reply string
// 			if err := client.Call("EzRpc.Main", args, &reply); err != nil {
// 				log.Println("call EzRpc.Main err:", args, err.Error())
// 				return
// 			}
// 			log.Println("!!r:", reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }

// func main() {
// 	var wg sync.WaitGroup
// 	typ := reflect.TypeOf(&wg)
// 	for i := 0; i < typ.NumMethod(); i++ {
// 		method := typ.Method(i)
// 		argv := make([]string, 0, method.Type.NumIn())
// 		returns := make([]string, 0, method.Type.NumOut())
// 		// 从1开始，第0个入参是wg自己
// 		for j := 1; j < method.Type.NumIn(); j++ {
// 			argv = append(argv, method.Type.In(j).Name())
// 		}
// 		for j := 0; j < method.Type.NumOut(); j++ {
// 			returns = append(returns, method.Type.Out(j).Name())
// 		}
// 		log.Printf("func (w %s) %s(%s) %s",
// 			typ.Elem().Name(),
// 			method.Name,
// 			strings.Join(argv, ","),
// 			strings.Join(returns, ","),
// 		)
// 	}
// }

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

// func startServer(addr chan string) {
// 	var foo Foo
// 	if err := ezrpc.Register(&foo); err != nil {
// 		log.Fatal("register err:", err.Error())
// 	}

// 	l, err := net.Listen("tcp", ":9977")
// 	if err != nil {
// 		log.Fatal("network err:", err.Error())
// 	}
// 	log.Println("start rpc server on:", l.Addr())
// 	addr <- l.Addr().String()
// 	ezrpc.Accept(l)
// }

// func startServer(addr chan string) {
// 	var foo Foo
// 	l, _ := net.Listen("tcp", ":9999")
// 	_ = ezrpc.Register(&foo)
// 	ezrpc.HandleHTTP()
// 	addr <- l.Addr().String()
// 	_ = http.Serve(l, nil)
// }

// func main() {
// 	log.SetFlags(log.Lshortfile | log.LstdFlags)
// 	addr := make(chan string)
// 	go startServer(addr)

// 	client, _ := ezrpc.Dial("tcp", <-addr)
// 	defer client.Close()

// 	time.Sleep(time.Second)
// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			args := &Args{Num1: i, Num2: i * i}
// 			var reply int
// 			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
// 				log.Fatal("call Foo.Sum error", err.Error())
// 			}
// 			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }

// func call(addrch chan string) {
// 	client, _ := ezrpc.DialHTTP("tcp", <-addrch)
// 	defer func() {
// 		_ = client.Close()
// 	}()

// 	time.Sleep(time.Second)
// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			args := &Args{Num1: i, Num2: i * i}
// 			var reply int
// 			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
// 				log.Fatal("call Foo.Sum error", err.Error())
// 			}
// 			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }

// func main() {
// 	log.SetFlags(log.Lshortfile | log.LstdFlags)
// 	ch := make(chan string)
// 	go call(ch)
// 	startServer(ch)
// }

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	err := http.Serve(l, nil)
	if err != nil {
		log.Println("start registry err:", err.Error())
	}
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := ezrpc.NewServer()
	_ = server.Register(&foo)
	wg.Done()
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	server.Accept(l)
}

// func startServer(registryAddr string) {
// 	var foo Foo
// 	l, _ := net.Listen("tcp", ":0")
// 	server := ezrpc.NewServer()
// 	_ = server.Register(&foo)
// 	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
// 	server.Accept(l)
// }

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(registry string) {
	d := xclient.NewEzRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// var i = 1
	// foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(registry string) {
	d := xclient.NewEzRegistryDiscovery(registry, 0)
	// d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1})

	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// var i = 1
	// foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)

	}
	wg.Wait()
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	registry := "http://localhost:9999/_ezrpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	// go startServer(registry)

	wg.Add(2)
	go startServer(registry, &wg)
	go startServer(registry, &wg)
	wg.Wait()

	time.Sleep(time.Second * 3)
	call(registry)
	// broadcast(registry)
}
