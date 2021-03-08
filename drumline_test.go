package drumline

import (
  "testing"
  "time"
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
  dl.Reset()
  time.Sleep(100 * time.Millisecond)
  if !x { t.Errorf("Progress should have been able to proceed (x)") }
  if !y { t.Errorf("Progress should have been able to proceed (y)") }
  dl.Close()
}
