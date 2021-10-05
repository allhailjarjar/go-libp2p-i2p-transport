package i2p

import (
	"context"

	tpt "github.com/libp2p/go-libp2p-core/transport"
)

type listener struct {
	ctx    context.Context
	cancel func()
}

func (l *listener) Accept() (tpt.CapableConn, error) {
	return nil, nil
}

func (l *listener) Close() error {
	return nil
}
