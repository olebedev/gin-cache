# gin-cache
Tiny and simple cache middleware for gin framework

## Usage

```go
package main

import (
  "time"

  "github.com/gin-gonic/gin"
  "github.com/olebedev/gin-cache"
)

func main() {
  r := gin.New()

  r.Use(cache.New(cache.Options{
    // set expire duration
    // by default zero, it means that cached content won't drop
    Expire: 5 * time.Minute,

    // store interface, see cache.go
    // by default it uses cache.InMemory
    Store: func() *cache.LevelDB {
      store, err := cache.NewLevelDB("cache")
      panicIf(err)
      return store
    }(),

    // it uses slice listed below as default to calculate key, if `Header` slice is not specified
    Header: []string{
			"User-Agent",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cookie",
			"User-Agent",
		},

    // *gin.Context.Abort() will be invoked immediately after cache has been served
    // so, you can change this, but you should manage c.Writer.Written() flag by self
    // example:
    // func config(c *gin.Context) {
    //   if c.Writer.Written() {
    //      return
    //   }
    //   // else serve content
    //   ...
    // }
    DoNotUseAbort: false,
  }))

  r.Run(":3000")
}
```


### TODO
- [x] inmemory store
- [x] leveldb store
- [ ] cache_test.go
- [ ] leveldb_test.go
- [ ] redis store
- [ ] memcache store
- [ ] add CI tool
