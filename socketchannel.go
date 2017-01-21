package socketchannel

import (
    "net"
    "sync"
    "bufio"
    "errors"
)

var (
    errNoCall  = errors.New("No call to receive response")
    errClosed  = errors.New("Closed")
)

type Response func(*bufio.Reader) ([]byte, error)

type Channel struct {
    addr string
    conn net.Conn
    callQueue []*Call
    cqm sync.Mutex
    rd *bufio.Reader
    resp Response
}

type Call struct {
    res []byte
    err error
    done chan *Call
}

func New(addr string, resp Response) (*Channel, error) {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return nil, err
    }
    ch := &Channel{
        addr: addr,
        conn: conn,
        callQueue: make([]*Call, 0),
        rd: bufio.NewReader(conn),
        resp: resp,
    }
    go ch.dispatch()
    return ch, nil
}

func (ch *Channel) Call(req []byte) ([]byte, error) {
    if ch.conn == nil {
        return nil, errClosed
    }
    _, err := ch.conn.Write(req)
    if err != nil {
        ch.wakeupAll(err)
        return nil, err
    }
    call := &Call {
        done: make(chan *Call, 1),
    }
    ch.cqm.Lock()
    ch.callQueue = append(ch.callQueue, call)
    ch.cqm.Unlock()
    call = <-call.done
    return call.res, call.err
}

func (ch *Channel) dispatch() {
    defer func() {
        if e:= recover(); e != nil {
            ch.wakeupAll(e.(error))
        }
    }()
    R := ch.rd
    resp := ch.resp
    for {
        r, err := resp(R)
        if err != nil {
            ch.wakeupAll(err)
            break
        }
        if len(ch.callQueue) == 0 {
            ch.wakeupAll(errNoCall)
            break
        }
        ch.cqm.Lock()
        call := ch.callQueue[0]
        ch.callQueue = ch.callQueue[1:]
        ch.cqm.Unlock()
        call.res = r
        call.done <- call
    }
}

func (ch *Channel) wakeupAll(err error) {
    ch.conn.Close()
    ch.conn = nil
    ch.cqm.Lock()
    for _, call := range(ch.callQueue) {
        call.err = err
        call.wakeup()
    }
    ch.cqm.Unlock()
}

func (ch *Channel) Close() {
    ch.wakeupAll(errClosed)
}

func (c *Call) wakeup() {
    select {
    case c.done <- c:
    default:
        panic("Call wakeup error")
    }
}
