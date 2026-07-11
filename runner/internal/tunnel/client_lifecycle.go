package tunnel

func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		c.lifecycleMu.Lock()
		c.stopped.Store(true)
		c.connected.Store(false)
		c.cancel()
		c.connMu.Lock()
		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}
		if c.dialConn != nil {
			_ = c.dialConn.Close()
			c.dialConn = nil
		}
		c.connMu.Unlock()
		c.lifecycleMu.Unlock()
		if c.dispatcher != nil {
			c.dispatcher.Close()
		}
		c.wg.Wait()
	})
}
