package drumline

import (
  "sync"
)


type Base struct {
  cond sync.Cond
  level int64
}

type Checkpoint struct{
  b *Base
  level int64
}
