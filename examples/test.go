package main

import (
    "github.com/lvshaco/gosocketchannel"
    "bufio"
    "log"
    "sync"
    "time"
    "bytes"
)

type WG struct {
    sync.WaitGroup
}
func (wg* WG) wrap(f func()) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        f()
    }()
}

func assert(cond bool, err error) {
    if !cond {
        panic(err)
    }
}

func main() {
    times := 100000
    ncurr := 1000
    perTimes := times/ncurr
    content := bytes.Repeat([]byte{49}, 1024)
    content = append(content, '\n')

    recv := 0
    wg := &WG{}
    sc, err := socketchannel.New("127.0.0.1:12345",
        func(rd *bufio.Reader)([]byte, error) {
            s, err := rd.ReadSlice('\n')
            s = s[:len(s)-1]
            //log.Println("recv:", recv)
            recv ++
            return s, err
    })
    assert(err == nil, err)

    work := func (n int) {
        for i:=0; i<perTimes; i++ {
            //log.Println("Call i:", i)
            _, err := sc.Call(content)//([]byte("hello world\n"))
            if err != nil {
                log.Println(err)
                break
            }
          //  log.Println("<= ", n, i, string(r))
          //  time.Sleep(time.Second*3)
        }
    }

    t1 := time.Now()
    for i:=0; i<ncurr; i++ {
        i:=i
        wg.wrap(func() {
            work(i)
        })
    }
    wg.Wait()
    t2 := time.Now()
    d := t2.Sub(t1)
    log.Printf("done: times=%d routine=%d", times, ncurr)
    log.Println(d)
}
