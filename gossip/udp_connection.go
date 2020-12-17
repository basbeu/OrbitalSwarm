package gossip

import (
	"bytes"
	"net"
	"time"

	"go.dedis.ch/onet/v3/log"
)

// stopMsg is used to notify the listener when we want to close the
// connection, so that the listener knows it can stop listening.
const stopMsg = "stop"

// UDPServer server
type UDPServer struct {
	Address *net.UDPAddr
	socket  *net.UDPConn

	listener       <-chan UDPPacket
	listenerClosed <-chan bool
	sender         chan<- UDPPacket
	senderClosed   <-chan bool

	close bool

	handlingFinished chan bool
}

// UDPPacket packet interface for UDPServer
type UDPPacket struct {
	data []byte
	addr *net.UDPAddr
}

// NewUDPServer create a new udp server
func NewUDPServer(addr string) (*UDPServer, error) {
	// Validate IP Address
	address, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	// UDP for gossipers
	socket, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Panic(err)
	}

	udpAddress := socket.LocalAddr()
	server := &UDPServer{
		Address:          udpAddress.(*net.UDPAddr),
		handlingFinished: make(chan bool),
		socket:           socket,
	}

	return server, nil
}

// Run Start the udp server
func (s *UDPServer) Run() (<-chan UDPPacket, chan<- UDPPacket, chan<- bool) {

	// Start udpSender and udpListener
	listener, listenerClosed := s.udpListener()
	sender, senderClosed := s.udpSender()

	s.listenerClosed = listenerClosed
	s.listener = listener
	s.senderClosed = senderClosed
	s.sender = sender

	return listener, sender, s.handlingFinished
}

// Stop the server
func (s *UDPServer) Stop() {
	s.close = true
	s.socket.WriteTo([]byte(stopMsg), s.Address)

	<-s.listenerClosed
	<-s.handlingFinished
	close(s.sender)

	<-s.senderClosed
	s.socket.Close()
}

func (s *UDPServer) udpListener() (<-chan UDPPacket, <-chan bool) {
	listener := make(chan UDPPacket, 1024)
	listeningClosed := make(chan bool)

	stopBytes := []byte(stopMsg)

	go func() {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("ERRROR %s", err)
			}
		}()
		for {
			buffer := make([]byte, 20480)

			s.socket.SetReadDeadline(time.Now().Add(2 * time.Second))

			length, src, _ := s.socket.ReadFromUDP(buffer)

			if bytes.Compare(buffer[:len(stopBytes)], stopBytes) == 0 || src == s.Address || s.close {
				// Close listener
				close(listener)
				close(listeningClosed)
				return
			} else if length == 0 {
				// Discard the message
			} else {
				listener <- UDPPacket{data: buffer[:length], addr: src}
			}
		}
	}()

	return listener, listeningClosed
}

func (s *UDPServer) udpSender() (chan<- UDPPacket, <-chan bool) {
	sending := make(chan UDPPacket, 1024)
	sendingClosed := make(chan bool)

	go func() {
		for {
			packet, ok := <-sending
			if ok {
				_, err := s.socket.WriteTo(packet.data, packet.addr)
				if err != nil {
					// Discard the message
					log.Printf("Discarded message while sending on socket")
				}
			} else {
				// Close sender
				s.socket.Close()
				close(sendingClosed)
				return
			}
		}
	}()
	return sending, sendingClosed
}
