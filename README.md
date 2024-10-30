# bucketmap

bucketmap提供了一种并发安全的map，将key分散到不同的bucket中，减小锁粒度。

## 示例

```go
package main

import (
	"fmt"

	"github.com/eachain/bucketmap"
)

func main() {
	m := bucketmap.Make[string, any]()
	m.Store("a", 123)
	m.Store("b", "xyz")

	a, _ := m.Load("a")
	fmt.Printf("a: %v\n", a)
	b, _ := m.Load("b")
	fmt.Printf("b: %v\n", b)
	// Output:
	// a: 123
	// b: xyz
}
```