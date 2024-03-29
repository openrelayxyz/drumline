// Package drumline provides a synchronization primitive for keeping multiple
// goroutines in lock step, within an established threshold.
package drumline

import (
  "time"
  "runtime"
  "sync"
  "sync/atomic"
  "math"
)

// Drumline ensures that goroutines stay in step within an acceptable
// threshold.
type Drumline struct {
  channels map[int]chan struct{}
  resets map[int]chan struct{}
  done map[int]struct{}
  resetCh chan struct{}
  scales map[int]int64
  minscale int64
  stepSize map[int]int64
  steps map[int]*int64
  doneCh chan int
  buffer int
  started bool
  closed bool
  quit chan struct{}
  lock sync.RWMutex
}

// NewDrumline creates a drumline. `buffer` indicates the maximum number of
// steps ahead one goroutine can be relative to the least advanced goroutine.
func NewDrumline(buffer int) *Drumline {
  return &Drumline{
    channels: make(map[int]chan struct{}),
    resets: make(map[int]chan struct{}),
    scales: make(map[int]int64),
    minscale: math.MaxInt64,
    stepSize: make(map[int]int64),
    steps: make(map[int]*int64),
    resetCh: make(chan struct{}),
    doneCh: make(chan int),
    done: make(map[int]struct{}),
    buffer: buffer,
    quit: make(chan struct{}),
  }
}

func (dl *Drumline) Add(i int) {
  dl.AddScale(i, 1)
}

// Add starts tracking a new goroutine in the Drumline
func (dl *Drumline) AddScale(i int, scale int64) {
  if scale < 1 {
    scale = 1 // Prevent divide by zero errors later, and I don't know what negative scale would mean.
  }
  dl.lock.Lock()
  dl.channels[i] = make(chan struct{}, dl.buffer)
  dl.resets[i] = make(chan struct{})
  dl.scales[i] = scale
  dl.steps[i] = new(int64)
  if scale < dl.minscale {
    dl.minscale = scale
    for j, s := range dl.scales {
      dl.stepSize[j] = s / dl.minscale
    }
  }
  dl.stepSize[i] = scale / dl.minscale
  dl.lock.Unlock()
  if !dl.started {
    dl.started = true
    go func() {
      for {
        for _, ch := range dl.channels {
          select {
          case <-ch:
          case thread := <-dl.doneCh:
            dl.lock.Lock()
            delete(dl.channels, thread)
            delete(dl.resets, thread)
            dl.done[thread] = struct{}{}
            dl.lock.Unlock()
            if len(dl.channels) == 0 {
              <-dl.quit
              return
            }
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

func (dl *Drumline) Done(i int) {
  dl.doneCh <- i
}

// Step advances a specific goroutine. If this would put that goroutine too far
// ahead of the rest of the drumline, this will block until other goroutines
// start to catch up.
func (dl *Drumline) Step(i int) {
  dl.lock.RLock()
  if _, ok := dl.done[i]; ok {
    dl.lock.RUnlock()
    return
  }
  atomic.AddInt64(dl.steps[i], 1)
  if *dl.steps[i] < dl.stepSize[i] {
    dl.lock.RUnlock()
    return
  }
  *dl.steps[i] = 0
  ch := dl.channels[i]
  rch := dl.resets[i]
  dl.lock.RUnlock()
  select {
  case ch <- struct{}{}:
  case <- rch:
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
      dl.lock.Lock()
      dl.channels[i] = make(chan struct{}, dl.buffer)
      dl.lock.Unlock()
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
