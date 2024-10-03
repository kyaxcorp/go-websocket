package websocket

import "github.com/kyaxcorp/go-helper/_context"

func (c *Client) GetCancelContext() *_context.CancelCtx {
	return c.ctx
}
