// Package drumline provides a synchronization primitive for keeping multiple
// goroutines in lock step, within an established threshold.
package drumline

import (
  "time"
  "runtime"
)

// Drumline ensures that goroutines stay in step within an acceptable
// threshold.
type Drumline struct {
  channels map[int]chan struct{}
  resets map[int]chan struct{}
  resetCh chan struct{}
  buffer int
  started bool
  closed bool
  quit chan struct{}
}

// NewDrumline creates a drumline. `buffer` indicates the maximum number of
// steps ahead one goroutine can be relative to the least advanced goroutine.
func NewDrumline(buffer int) *Drumline {
  return &Drumline{
    channels: make(map[int]chan struct{}),
    resets: make(map[int]chan struct{}),
    resetCh: make(chan struct{}),
    buffer: buffer,
    quit: make(chan struct{}),
  }
}

// Add starts tracking a new goroutine in the Drumline
func (dl *Drumline) Add(i int) {
  dl.channels[i] = make(chan struct{}, dl.buffer)
  dl.resets[i] = make(chan struct{})
  if !dl.started {
    dl.started = true
    go func() {
      for {
        for _, ch := range dl.channels {
          select {
          case <-ch:
          case <-dl.resetCh:
            for _, v := range dl.resets {
              select {
              case v <- struct{}{}:
              default:
              }
              runtime.Gosched()
            }
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
  select {
  case dl.channels[i] <- struct{}{}:
  case <- dl.resets[i]:
  }
}


// Reset returns a channel that will produce a message after the specified
// duration. If the message is consumed, the drumline will reset; if the
// message is not consumed the reset will expire. If the drumline is closed,
// the returned channel will be nil, and will never produce a message.
func (dl *Drumline) Reset(d time.Duration) chan time.Time {
  if dl.closed { return nil }
  ret := make(chan time.Time)
  go func(ret chan time.Time) {
    t := <- time.After(d)
    select {
    case ret <- t:
    default:
      return
    }
    for i := range dl.channels {
      dl.channels[i] = make(chan struct{}, dl.buffer)
    }
    select {
    case dl.resetCh <- struct{}{}:
    default:
    }
  }(ret)
  return ret
}

// Close cleans up the drumline. Close() must be called when a Drumline is
// discarded, or it will not be garbage collected.
func (dl *Drumline) Close() {
  dl.closed = true
  dl.quit <- struct{}{}
  close(dl.quit)
}
