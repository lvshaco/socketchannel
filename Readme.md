A simple socketchannel
=====
针对单条socket连接的非阻塞，并发读写
类似于net/rpc，但是更加简单，纯粹，仅仅提供非阻塞并发的写法
具体用法看examples
```
package main

import (
    "github.com/lvshaco/gosocketchannel"
    "bufio"
    "log"
    "sync"
    "time"
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
    wg := &WG{}
    sc, err := socketchannel.New("127.0.0.1:12345",
        func(rd *bufio.Reader)([]byte, error) {
            s, err := rd.ReadSlice('\n')
            s = s[:len(s)-1]
            return s, err
    })
    assert(err == nil, err)

    work := func (i int) {
        for {
            r, err := sc.Call([]byte("hello world\n"))
            if err != nil {
                log.Println(err)
                break
            }
            log.Println(i, string(r))
            time.Sleep(time.Second*3)
        }
    }

    for i:=0; i<10; i++ {
        i:=i
        wg.wrap(func() {
            work(i)
        })
    }
    wg.Wait()
    log.Println("done")
}
```
ps: golang语言习作
