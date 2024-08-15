package proxy

import "sync"

type weight uint8

const High weight = 3
const Medium weight = 2
const Low weight = 1

type Instance interface {
	GetAddr() string
	AddConn()
	RemoveConn()
}

type instance struct {
	addr                  string
	connections, maxConns int
	weight                uint8
	mu                    sync.RWMutex
}

func InitInstance(addr string, maxConns int, levelW weight) *instance {
	return &instance{

		addr:        addr,
		connections: 0,
		weight:      uint8(levelW),
		maxConns:    maxConns,
		mu:          sync.RWMutex{},
	}
}

func (s *instance) AddConn() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections >= s.maxConns {
		return
	}

	s.connections++
}

func (s *instance) RemoveConn() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections <= 0 {
		return
	}

	s.connections--
}

func (s *instance) GetAddr() string {
	return s.addr
}

func (s *instance) calcWorkLoadPerc() int {

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.connections / s.maxConns * 100

}

func (s *instance) getWeight() uint8 {
	return s.weight
}

type IBalancer interface {
	GetInstance(service string) Instance
}

type Balancer struct {
	services map[string][]*instance
	mu       sync.RWMutex
}

func NewBalancer(maxServicese int) *Balancer {

	return &Balancer{
		services: make(map[string][]*instance, maxServicese),
		mu:       sync.RWMutex{},
	}

}

func (b *Balancer) AddService(service string) {

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.services[service]; ok {
		return
	}

	b.services[service] = make([]*instance, 0)

}

func (b *Balancer) AddInstance(service string, i *instance) {

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.services[service]; !ok {
		return
	}

	b.services[service] = append(b.services[service], i)

}

func (b *Balancer) GetInstance(service string) Instance {

	i := b.services[service][0]

	for s := 1; s < len(b.services[service]); s++ {
		if i.calcWorkLoadPerc() > b.services[service][s].calcWorkLoadPerc() {
			i = b.services[service][s]
		}
	}

	i.AddConn()

	return i
}
