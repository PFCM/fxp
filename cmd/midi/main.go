// command midi checks that midi is working.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/pfcm/fxp/midi"
	"github.com/pfcm/fxp/midi/coremidi"
)

func main() {
	ctx := interruptContext()

	d := midi.Listen(ctx, coremidi.ReceiveAll)
	c := d.Subscribe()
	for m := range c {
		fmt.Println(m)
	}

	log.Println("all done")
}

func interruptContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}
