package ableton_js

import (
    "fmt"
    "log"
    "net"
    "strings"
    "time"
)

const (
    host            = "127.0.0.1"
    listenPort      = 9031 // specific port to always bind to on our side
    sendPort        = 9041 // remote Python UDP server
    udp4            = "udp4"
    sendBufferDepth = 128         // arbitrary value to absorb bursts
    writeDeadline   = time.Second // arbitrary value to avoid blocking forever
    readDeadline    = time.Second // arbitrary value to avoid blocking forever
)

type Connection struct {
    conn      *net.UDPConn
    stop      chan bool
    sendQueue chan []byte
    handler   func([]byte) error
    response  chan []byte
}

func NewConnection(receiveHandler func([]byte) error) *Connection {
    c := Connection{
        stop:      make(chan bool),
        sendQueue: make(chan []byte, sendBufferDepth),
        handler:   receiveHandler,
        response:  make(chan []byte),
    }

    return &c
}

func (c *Connection) receiver() {
    var buffer []byte

    for {
        select {
        case <-c.stop:
            return
        default:
        }

        buffer = make([]byte, 65535)

        _ = c.conn.SetReadDeadline(time.Now().Add(readDeadline))
        n, _, err := c.conn.ReadFromUDP(buffer)
        if err != nil {
            if strings.Contains(err.Error(), "i/o timeout") {
                continue
            }

            log.Printf("warning: c.conn.Read failed because %v", err)
        }

        if n == 0 {
            continue
        }

        c.response <- HandleResponse(buffer[0:n])
    }
}

func (c *Connection) Open() error {
    if c.conn != nil {
        return fmt.Errorf("cannot open; connection is not nil")
    }

    localAddr, err := net.ResolveUDPAddr(udp4, fmt.Sprintf("%v:%v", host, listenPort))
    if err != nil {
        return err
    }

    remoteAddr, err := net.ResolveUDPAddr(udp4, fmt.Sprintf("%v:%v", host, sendPort))
    if err != nil {
        return err
    }

    conn, err := net.DialUDP(udp4, localAddr, remoteAddr)
    if err != nil {
        return err
    }

    c.conn = conn

    go c.receiver()

    return nil
}

func (c *Connection) Send(payload []byte) error {
    if c.conn == nil {
        return fmt.Errorf("cannot send; connection is nil")
    }

    _ = c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
    _, err := c.conn.Write(payload)

    return err
}

func (c *Connection) GetCurrentSongTime() ([]byte, error) {
    err := c.Send(GetCurrentSongTime())
    if err != nil {
        return []byte{}, nil
    }

    return <-c.response, nil
}

func (c *Connection) Close() {
    if c.conn == nil {
        return
    }

    _ = c.conn.Close()
    c.conn = nil

    select {
    case c.stop <- true:
    default:
    }
}
