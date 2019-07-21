package connection

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type OnConnectCallback func(conn net.Conn) error

type Info struct {
	Net     string
	Address string
	TLS     *tls.Config
}

type manager struct {
	sync.Mutex
	info                      *Info
	conn                      net.Conn
	lastMasterConnectionError error
	shouldStop                bool
	onConnectCallbacks        map[int]OnConnectCallback
	maxFailureAttempts        int
	tokenCounter              int
}

func (m *manager) Get() (net.Conn, error) {
	if m.conn != nil && m.lastMasterConnectionError == nil {
		return m.conn, nil
	}
	attempts := -1
	for !m.shouldStop {

		if m.maxFailureAttempts > 0 && attempts == m.maxFailureAttempts {
			return nil, m.lastMasterConnectionError
		}

		if m.info.TLS != nil {
			m.conn, m.lastMasterConnectionError = tls.Dial(m.info.Net, m.info.Address, m.info.TLS)
		} else {
			m.conn, m.lastMasterConnectionError = net.Dial(m.info.Net, m.info.Address)
		}

		if m.lastMasterConnectionError != nil {
			log.Println("could not connect to", m.info.Address)
			attempts++
			<-time.After(time.Second * 2)
			log.Printf("trying to connect to %s@%s, attempt=%d\n", m.info.Net, m.info.Address, attempts)
		} else {
			return m.conn, m.onConnect()
		}
	}
	return nil, errors.New("failed to connect, manager closed")
}

func (m *manager) SetMaxFailureAttempts(maxFailureAttempts int) {
	m.maxFailureAttempts = maxFailureAttempts
}

func (m *manager) Close() error {
	if err := m.conn.Close(); err != nil {
		return err
	}

	m.shouldStop = true
	return nil
}

func (m *manager) onConnect() error {
	m.Lock()
	defer m.Unlock()

	for _, c := range m.onConnectCallbacks {
		err := c(m.conn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *manager) AddOnConnectCallback(callback OnConnectCallback) int {
	m.Lock()
	defer m.Unlock()

	m.tokenCounter++
	m.onConnectCallbacks[m.tokenCounter] = callback
	return m.tokenCounter
}

func (m *manager) RemoveOnConnectCallback(key int) {
	m.Lock()
	defer m.Unlock()
	delete(m.onConnectCallbacks, key)
}

func NewManager(info *Info) *manager {
	return &manager{
		info: info,
	}
}
