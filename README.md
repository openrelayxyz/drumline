# Drumline

Drumline provides a synchronization primitive for keeping multiple goroutines
in lock step.

The original use case was a process that consumed a Kafka topic consisting of
multiple partitions. If the goroutine processing one partition got too far
ahead or behind the goroutines processing other partitions, it could cause
problems if the process was disrupted and needed to be resumed later. Drumline
was created to ensure that no process could get too far ahead of the others.

## Example

```
package main

import
  (
    "fmt"
    "github.com/openrelayxyz/drumline"
  )

func main (){
  dl := drumline.NewDrumline(1)

  for i := 0 ; i < 3; i++ {
    dl.Add(1)
    go func() {
      for j := 0; j < 10; j++ {
        fmt.Printf("%v: %v", i, j)
        dl.Step(i)
      }
    }()
  }
}


```
