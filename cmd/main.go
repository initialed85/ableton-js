package main

import (
    "github.com/initialed85/ableton-js/pkg/ableton_js"
    "log"
    "os"
    "os/signal"
    "time"
)

func waitForCtrlC() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    for {
        _ = <-c
        break
    }
}

func main() {
    handler := func(payload []byte) error {
        log.Printf("payload=%#+v", payload)

        return nil
    }

    connection := ableton_js.NewConnection(handler)

    err := connection.Open()
    if err != nil {
        log.Fatal(err)
    }

    defer connection.Close()

    t := time.NewTicker(time.Millisecond * 100)

    go func() {
        for {
            <-t.C

            before := time.Now()

            payload, err := connection.GetCurrentSongTime()
            if err != nil {
                log.Fatal(err)
            }

            after := time.Now()

            log.Printf("%v to get %v", after.Sub(before), string(payload))
        }
    }()

    waitForCtrlC()
}
