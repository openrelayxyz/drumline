// Package drumline provides a synchronization primitive for keeping multiple
// goroutines in lock step, within an established threshold.
package drumline

// Drumline ensures that goroutines stay in step within an acceptable
// threshold.
type Drumline struct {
  channels map[int]chan struct{}
  buffer int
  started bool
  quit chan struct{}
}

// NewDrumline creates a drumline. `buffer` indicates the maximum number of
// steps ahead one goroutine can be relative to the least advanced goroutine.
func NewDrumline(buffer int) *Drumline {
  return &Drumline{
    channels: make(map[int]chan struct{}),
    buffer: buffer,
    quit: make(chan struct{}),
  }
}

// Add starts tracking a new goroutine in the Drumline
func (dl *Drumline) Add(i int) {
  dl.channels[i] = make(chan struct{}, dl.buffer)
  if !dl.started {
    dl.started = true
    go func() {
      for {
        for _, ch := range dl.channels {
          select {
          case <-ch:
          case <-dl.quit:
            return
          }
        }
      }
    }()
  }
}

// Step advances a specific goroutine. If this would put that goroutine too far
// ahead of the rest of the drumline, this will block until other goroutines
// start to catch up.
func (dl *Drumline) Step(i int) {
  dl.channels[i] <- struct{}{}
}

// Close cleans up the drumline. Close() must be called when a Drumline is
// discarded, or it will not be garbage collected.
func (dl *Drumline) Close() {
  dl.quit <- struct{}{}
  close(dl.quit)
}
