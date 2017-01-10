/*
Discovery program that uses goroutines and channels - just to play with them.

I create a bunch of objects, each of which run a CPU intensive task; thus making it
obvious whether real parallelism is being achieved.
Tasks are started by receiving a start token on their dedicated channel.
Once running, tasks report progress regularly, then indicate done using a STOP token.
A separate goroutine validates the tokens on the channels.
A wait group is used to wait for all the token validation to complete.
Invalid tokens cause a panic.

Usage of /home/gavin/work/go/workspace/bin/cpu_load:
  -b int
    	number of iterations per gor (default 1000000000)
  -n int
    	number of goroutines to launch (default 24)

*/
package main


import (
    "fmt"
    "math"
    "flag"
    "time"
)

const DEFAULT_BIG = 1000 * 1000 * 1000
const DEFAULT_NGOR = 24
const BUFFER = 4

// allowed messages
const START = "Start"
const STOP = "Stop"
const PROGRESS = "InProgress"

// Start of Object thing
type thing float64

// not using pointer as want a local copy of the thing only
func (t thing) run(me int, big int, two_chan chan string, cntl_chan chan string) {
    waituntil := <-cntl_chan
    t.validate_start(waituntil)
    for n:=0; n<big; n++ {
        newt := thing(math.Sqrt(float64(t * t + thing(me))))
        if n % (big / 4) == 0 {
            two_chan <- fmt.Sprintf("%s %d %d\n", PROGRESS, me, n)
            fmt.Printf("%s %d %d\n", PROGRESS, me, n)
        }
        t = newt
    }
    fmt.Printf("%d ended with final result %f\n", me, float64(t))
    two_chan <- STOP
}

func (t thing) validate_start(message string) {
    if message != START {
        panic(fmt.Sprintf("Unexpected start message = %s", message))
    }
}
// end of thing


func validate_stream(data_chan chan string, wait_chan chan int) {
    for input := <-data_chan; input != STOP; input = <-data_chan {
        if input[:len(PROGRESS)] == PROGRESS {
            fmt.Println(input)
        } else {
            panic(fmt.Sprintf("Unknown message %s\n", input))
        }
    }
    wait_chan <- 0
}

func report(s string) {
    fmt.Printf("%s : %s\n", time.Now().Local().String(), s)
}

func main() {
    ngor := flag.Int("n", DEFAULT_NGOR, "number of goroutines to launch")
    bignumb := flag.Int("b", DEFAULT_BIG, "number of iterations per gor")
    flag.Parse()
    data_comms := make([]chan string, 0)
    cntl_comms := make([]chan string, 0)
    report(fmt.Sprintf("Setting things up for %d gors", *ngor))
    for n := 0; n < *ngor; n++ {
        data_comms = append(data_comms, make(chan string, BUFFER))
        cntl_comms = append(cntl_comms, make(chan string, BUFFER))
    }
    gors := make([]thing, 0)
    for n := 0; n < *ngor; n++ {
        gors = append(gors, thing(1.0 * n))
    }
    for n, t := range(gors) {
        go t.run(n, *bignumb, data_comms[n], cntl_comms[n])
    }
    report("Starting the gors")
    for n, _ := range(gors) {
        cntl_comms[n] <- START
    }
    report("All started - waiting for them to finish")
    /*
    var wg sync.WaitGroup
    for n, _ := range(gors) {
        wg.Add(1)
        go validate_stream(data_comms[n], &wg)
    }
    wg.Wait()
    */
    wait_chan := make(chan int)
    for n, _ := range(gors) {
        go validate_stream(data_comms[n], wait_chan)
    }
    for i := 0; i < len(gors); i++ {
        <-wait_chan
    }
    report("All done")
}
