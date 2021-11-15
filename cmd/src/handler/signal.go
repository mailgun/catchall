package handler

import "os"
import "os/signal"
import "syscall"

type SigInt = chan os.Signal

func NewSigInt() SigInt {
    ret := make(SigInt)
    signal.Notify(ret,
        syscall.SIGINT,
        syscall.SIGQUIT,
        syscall.SIGTERM,
    )
    return ret
}

func WaitCloseSigInt(self SigInt) {
    select {
    case <-self: close(self)
    }
}

func CloseSigInt(self SigInt) {
    close(self)
}
