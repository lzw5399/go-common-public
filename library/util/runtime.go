package util

import "runtime"

func FuncName() string {
    pc := make([]uintptr, 1)
    runtime.Callers(2, pc)
    f := runtime.FuncForPC(pc[0])

    return f.Name()
}
