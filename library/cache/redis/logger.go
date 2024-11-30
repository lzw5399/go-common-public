package fredis

import "context"

type NothingLogAdaptor struct {
}

func (n NothingLogAdaptor) Printf(ctx context.Context, format string, v ...interface{}) {
}
