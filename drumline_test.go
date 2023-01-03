package drumline

import (
  "testing"
  "time"
  "log"
)

func TestProgresLimiter(t *testing.T) {
  dl := NewDrumline(5)
  dl.Add(0)
  dl.Add(1)
  dl.Add(2)
  x := false
  y := false
  go func() {
    for i := 0; i < 7; i++ { dl.Step(0) }
    x = true
  }()
  go func() {
    for i := 0; i < 7; i++ { dl.Step(1) }
    y = true
  }()
  time.Sleep(100 * time.Millisecond)
  if x { t.Errorf("Progress should have been limited (x)") }
  if y { t.Errorf("Progress should have been limited (y)") }
  dl.Step(2)
  dl.Step(2)
  time.Sleep(100 * time.Millisecond)
  if !x { t.Errorf("Progress should have been able to proceed (x)") }
  if !y { t.Errorf("Progress should have been able to proceed (y)") }
  dl.Close()
}


func TestReset(t *testing.T) {
  dl := NewDrumline(5)
  dl.Add(0)
  dl.Add(1)
  dl.Add(2)
  x := false
  y := false
  go func() {
    for i := 0; i < 7; i++ { dl.Step(0) }
    x = true
  }()
  go func() {
    for i := 0; i < 7; i++ { dl.Step(1) }
    y = true
  }()
  time.Sleep(100 * time.Millisecond)
  if x { t.Errorf("Progress should have been limited (x)") }
  if y { t.Errorf("Progress should have been limited (y)") }
  <-dl.Reset(1 * time.Millisecond)
  time.Sleep(100 * time.Millisecond)
  if !x { t.Errorf("Progress should have been able to proceed (x)") }
  if !y { t.Errorf("Progress should have been able to proceed (y)") }
  dl.Close()
  if x := dl.Reset(0); x != nil {
    t.Errorf("Reset on closed drumline should return nil channel, got %v", x)
  }
}

func TestProgresLimiterScaled(t *testing.T) {
  dl := NewDrumline(5)
  dl.AddScale(0, 1)
  dl.AddScale(1, 10)
  dl.AddScale(2, 100)
  x := false
  y := false
  go func() {
    for i := 0; i < 7; i++ { dl.Step(0) }
    x = true
  }()
  go func() {
    for i := 0; i < 70; i++ { dl.Step(1) }
    y = true
  }()
  time.Sleep(100 * time.Millisecond)
  if x { t.Errorf("Progress should have been limited (x)") }
  if y { t.Errorf("Progress should have been limited (y)") }
  log.Printf("0: %v, 1: %v, 2: %v", *dl.steps[0], *dl.steps[1], *dl.steps[2])
  for i := 0; i < 700; i++ {
    dl.Step(2)
  }
  time.Sleep(100 * time.Millisecond)
  if !x { t.Errorf("Progress should have been able to proceed (x)") }
  if !y { t.Errorf("Progress should have been able to proceed (y)") }
  dl.Close()
}
func TestProgresLimiterScaledFlipped(t *testing.T) {
  dl := NewDrumline(5)
  dl.AddScale(0, 1)
  dl.AddScale(1, 10)
  dl.AddScale(2, 100)
  x := false
  y := false
  go func() {
    for i := 0; i < 700; i++ { dl.Step(2) }
    x = true
  }()
  go func() {
    for i := 0; i < 70; i++ { dl.Step(1) }
    y = true
  }()
  time.Sleep(100 * time.Millisecond)
  if x { t.Errorf("Progress should have been limited (x)") }
  if y { t.Errorf("Progress should have been limited (y)") }
  log.Printf("0: %v, 1: %v, 2: %v", *dl.steps[0], *dl.steps[1], *dl.steps[2])

  for i := 0; i < 7; i++ {
    dl.Step(0)
  }
  time.Sleep(100 * time.Millisecond)
  if !x { t.Errorf("Progress should have been able to proceed (x)") }
  if !y { t.Errorf("Progress should have been able to proceed (y)") }
  dl.Close()
}
