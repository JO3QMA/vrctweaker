package tray

// stopWaiterSet tracks Stop() callers waiting for the tray session to finish.
// complete broadcasts the same result to every registered waiter.
type stopWaiterSet struct {
	waiters []chan error
}

func (s *stopWaiterSet) add() chan error {
	ch := make(chan error, 1)
	s.waiters = append(s.waiters, ch)
	return ch
}

func (s *stopWaiterSet) complete(err error) {
	for _, ch := range s.waiters {
		ch <- err
	}
	s.waiters = nil
}
