package websocket

import (
	"context"

	"github.com/kyaxcorp/go-helper/_context"
)

func (h *Hub) SetContext(ctx context.Context) {
	if ctx == nil {
		ctx = _context.GetDefaultContext()
	}
	h.parentCtx = ctx

	h.NewCancelContext()
}

func (h *Hub) NewCancelContext() *Hub {
	h.ctx = _context.WithCancel(h.parentCtx)
	return h
}
