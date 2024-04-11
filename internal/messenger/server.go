package messenger

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Server struct {
	Server         net.Listener
	WelcomeMsg     string
	connections    map[string]net.Conn
	MaxConnections int
	allMessages    string
	messages       chan Message
	leaving        chan Message
	mutex          sync.Mutex
	Count          int
}

func NewServer(protocol, addres string, MaxConn int) (*Server, error) {
	ln, err := net.Listen(protocol, addres)
	if err != nil {
		return nil, err
	}
	con := make(map[string]net.Conn, MaxConn)
	messages := make(chan Message)
	leaving := make(chan Message)
	welcome, err := os.ReadFile("resources/welcomeMsg.txt")
	if err != nil {
		return nil, err
	}
	return &Server{
		Server:         ln,
		WelcomeMsg:     string(welcome),
		connections:    con,
		MaxConnections: MaxConn,
		messages:       messages,
		leaving:        leaving,
	}, nil
}

func (s *Server) client(conn net.Conn) {
	name, err := s.connect(conn)
	defer s.closeConnection(conn, name)
	if err != nil {
		return
	}
	input := bufio.NewScanner(conn)
	s.mutex.Lock()
	conn.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format(dateFormat), name)))
	s.mutex.Unlock()
	for input.Scan() {
		s.newMsg(&Message{
			Text: input.Text(),
			Time: time.Now().Format(dateFormat),
			User: name,
		})
		conn.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format(dateFormat), name)))
	}
}

func (s *Server) connect(conn net.Conn) (string, error) {
	s.mutex.Lock()
	s.Count++
	log.Println(s.Count)
	if s.Count > s.MaxConnections {
		conn.Write([]byte(FullRoomMsg))
		s.mutex.Unlock()
		return "", errors.New("Max connections")
	}
	s.welcome(conn)
	s.mutex.Unlock()
	name := s.newUserName(conn)
	conn.Write([]byte(s.allMessages))
	if name != "" {
		s.newUserNotification(name)
	}
	s.addConnection(conn, name)
	return name, nil
}

func (s *Server) welcome(conn net.Conn) {
	conn.Write([]byte(s.WelcomeMsg))
}

func (s *Server) newUserNotification(name string) {
	s.newMsg(&Message{
		Text: name + WelcomeMsg,
		Time: "",
		User: name,
	})
}

func (s *Server) leaveNotification(name string) {
	s.newMsg(&Message{
		Text: name + LeaveMsg,
		Time: "",
		User: name,
	})
}

func (s *Server) newUserName(conn net.Conn) string {
	input := bufio.NewScanner(conn)

	conn.Write([]byte(NAME))
	for input.Scan() {
		if input.Text() == "" {
			conn.Write([]byte(EmptyNameMsg))
			conn.Write([]byte(NAME))
			continue
		}
		if _, exist := s.connections[input.Text()]; exist {
			conn.Write([]byte(UsedNameMsg))
			conn.Write([]byte(NAME))
		} else {
			return input.Text()
		}
	}
	return input.Text()
}

func (s *Server) addConnection(conn net.Conn, name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if name != "" {
		s.connections[name] = conn
		log.Printf("User %s has joined, number of users:%d", name, s.Count)
	}
}

func (s *Server) closeConnection(conn net.Conn, name string) {
	log.Println("closing connection..")
	s.mutex.Lock()
	conn.Close()
	s.Count--
	delete(s.connections, name)
	log.Printf("User %s has left, number of users:%d", name, s.Count)
	s.mutex.Unlock()
	if name != "" {
		s.leaveNotification(name)
	}
}

func (s *Server) newMsg(msg *Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if msg.Text == "" {
		return
	}
	s.messages <- *msg
	if msg.Time != "" {
		s.allMessages += msg.string()
	}
}

func (s *Server) broadcaster() {
	for {
		select {
		case msg := <-s.messages:
			message := msg.string()
			s.mutex.Lock()
			for user, conn := range s.connections {
				if user != msg.User {
					conn.Write([]byte("\n" + message))
					conn.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format(dateFormat), user)))
				}
			}
			s.mutex.Unlock()
		case msg := <-s.leaving:
			message := msg.string()
			s.mutex.Lock()
			for user, conn := range s.connections {
				if user != msg.User {
					conn.Write([]byte("\n" + message))
					conn.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format(dateFormat), user)))
				}
			}
			s.mutex.Unlock()

		}
	}
}

func (s *Server) Start() {
	go s.broadcaster()
	defer s.Server.Close()
	for {
		conn, err := s.Server.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}

		go s.client(conn)
	}
}
