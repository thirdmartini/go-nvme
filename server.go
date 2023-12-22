package nvme

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/thirdmartini/go-nvme/internal/sys"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
)

var serialAccess sync.Mutex

type SessionInfo struct {
	Source string
	Ctrl   *Controller
}

type Server struct {
	Address     string
	SubSystems  map[string]Subsystem
	SessionInfo map[string]SessionInfo
	Lock        sync.Mutex
	listen      net.Listener

	wg   sync.WaitGroup
	quit chan bool

	debugLevel uint64
}

func (s *Server) RegisterSession(name string, ctrl *Controller) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	s.SessionInfo[name] = SessionInfo{
		Source: name,
		Ctrl:   ctrl,
	}
}

func (s *Server) UnRegisterSession(name string) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	delete(s.SessionInfo, name)
}

func (s *Server) GetSessions() []SessionInfo {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	sil := make([]SessionInfo, 0, len(s.SessionInfo))
	for k := range s.SessionInfo {
		sil = append(sil, s.SessionInfo[k])
	}
	return sil
}

func (s *Server) AddSubSystem(subsys Subsystem) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.SubSystems[subsys.GetNQN()] = subsys
}

func (s *Server) GetSubSystem(nqn string) Subsystem {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	subsys, ok := s.SubSystems[nqn]
	if !ok {
		return nil
	}
	return subsys
}

func (s *Server) ListSubSystems() []Subsystem {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	ls := make([]Subsystem, 0, len(s.SubSystems))

	for k := range s.SubSystems {
		ls = append(ls, s.SubSystems[k])
	}

	return ls
}

func (s *Server) RemoveSubSystem(nqn string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	subsys, ok := s.SubSystems[nqn]
	if !ok {
		return fmt.Errorf("nqn:%s does not exists", nqn)
	}
	delete(s.SubSystems, subsys.GetNQN())

	for id := range s.SessionInfo {
		session := s.SessionInfo[id]
		if session.Ctrl.Subsystem != subsys {
			continue
		}

		session.Ctrl.Close()
	}
	return nil
}

func (s *Server) startController(conn *sys.Conn) {
	sessionId := conn.RemoteAddr().String()

	ctrl := Controller{
		REGCtrlCaps: uint64(1)<<37 | uint64(15)<<24 | (uint64(protocol.NVMECtrlAttrMaxQueueSize) - 1),
		//		REGCtrlStatus: 0x1,              // Controller is ready
		REGCtrlStatus: 0x0,
		Version:       1<<16 | 3<<8 | 0, // 1.3.0
		ControllerID:  0,                // was 1
		Subsystem:     &InitSubsystem{},
		QueueID:       0,
		QueueSize:     protocol.NVMECtrlAttrMaxQueueSize, // default queue size
		Queue:         make([]NVMERequest, protocol.NVMECtrlAttrMaxQueueSize, protocol.NVMECtrlAttrMaxQueueSize),
		SQCUR:         0,
		SQHD:          0,
		Server:        s,

		SessionID: sessionId,

		FlowControlDisabled: false,
		waiting:             make(chan *NVMERequest, protocol.NVMECtrlAttrMaxQueueSize),
		completions:         make(chan *NVMERequest, protocol.NVMECtrlAttrMaxQueueSize),
	}

	for i, _ := range ctrl.Queue {
		req := &ctrl.Queue[i]
		ctrl.waiting <- req
	}

	if s.debugLevel == 0 {
		ctrl.Log = &tracer.NullTracer{}
	} else {
		ctrl.Log = &tracer.LogTracer{
			Id:    sessionId,
			Level: s.debugLevel,
		}
	}
	ctrl.Log.Begin(tracer.TraceController, "Session Started")

	s.RegisterSession(sessionId, &ctrl)
	defer func() {
		s.UnRegisterSession(sessionId)
		ctrl.Log.Trace(tracer.TraceController, "Session Terminated")
		conn.Close()
	}()

	if true {
		ctrl.Serve(conn)
		return
	} else {
		panic(false)
		//ctrl.ServeDeprecated(conn)
	}
}

func (s *Server) Serve() error {
	defer s.wg.Done()
	for {
		conn, err := s.listen.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
				log.Println("accept error", err)
			}
		} else {
			sconn := sys.NewConn(conn)
			fmt.Printf("Server Starded Queue\n")
			s.wg.Add(1)
			go func() {
				s.startController(sconn)
				s.wg.Done()
			}()
		}
	}
}

func (s *Server) Close() error {
	close(s.quit)
	s.listen.Close()
	fmt.Printf("Listen Closed\n")
	s.wg.Wait()
	fmt.Printf("Server Closed\n")
	return nil
}

func (s *Server) SetDebugLevel(level uint64) {
	s.debugLevel = level
}

func New(addr string) (*Server, error) {
	listen, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		SubSystems:  make(map[string]Subsystem),
		SessionInfo: make(map[string]SessionInfo),
		Address:     addr,
		listen:      listen,
		quit:        make(chan bool),
	}
	s.wg.Add(1)
	// We always have a discovery subsystem
	discovery := &DiscoverySubsystem{
		Server: s,
	}
	s.AddSubSystem(discovery)
	return s, nil
}
