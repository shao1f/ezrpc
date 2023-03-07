package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type EZRegister struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_ezrpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *EZRegister {
	return &EZRegister{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultEzRegister = New(defaultTimeout)

func (r *EZRegister) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now()
	}
	// fmt.Println("putServer succ:", addr)
}

func (r *EZRegister) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alives []string
	for alive, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alives = append(alives, alive)
		} else {
			delete(r.servers, alive)
		}
	}
	sort.Strings(alives)
	// fmt.Println("alive servers:", alives)
	return alives
}

func (r *EZRegister) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("X-EzRpc-Servers", strings.Join(r.aliveServers(), ","))
	case http.MethodPost:
		addr := req.Header.Get("X-EzRpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *EZRegister) HandleHTTP(registerPath string) {
	http.Handle(registerPath, r)
	log.Println("rpc register path:", registerPath)
}

func HandleHTTP() {
	DefaultEzRegister.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Microsecond
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	// X-EzRpc-Server 需要作为const，写错一个地方就GG
	req.Header.Set("X-EzRpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("ezrpc server: heart beat err:", err)
		return err
	}
	return nil
}
