package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const KEY_PREFIX = "gin:cache:"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Cached struct {
	status   int
	body     []byte
	header   http.Header
	expireAt time.Time
}

type Store interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
	Remove(string) error
	Update(string, []byte) error
	Keys() []string
}

type Options struct {
	Store  Store
	Expire time.Duration
}

func (o *Options) init() {}

type Cache struct {
	Store
	options Options
	expires map[string]time.Time
}

type wrappedWriter struct {
	gin.ResponseWriter
	body []byte
}

func (rw *wrappedWriter) Write(body []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(body)
	if err == nil {
		rw.body = body
	}
	return n, err
}

func New(o ...Options) gin.HandlerFunc {
	opts := Options{
		Store:  NewInMemory(),
		Expire: 0,
	}

	for _, i := range o {
		opts = i
		break
	}
	opts.init()

	cache := Cache{
		Store:   opts.Store,
		options: opts,
		expires: make(map[string]time.Time),
	}

	return func(c *gin.Context) {

		// only GET method available for caching
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		var cached Cached
		key := KEY_PREFIX + md5String(c.Request.URL.RequestURI())

		if data, err := cache.Get(key); err == ErrNotFound {
			// cache miss
			rw := wrappedWriter{ResponseWriter: c.Writer}
			c.Writer = &rw
			c.Next()
			cached = Cached{
				status: rw.Status(),
				body:   rw.body,
				header: rw.Header(),
				expireAt: func() time.Time {
					if cache.options.Expire == 0 {
						return time.Time{}
					} else {
						return time.Now().Add(cache.options.Expire)
					}
				}(),
			}

			var b bytes.Buffer
			enc := gob.NewEncoder(&b)

			// TODO: handle errors
			enc.Encode(cached)
			cache.Set(key, b.Bytes())

			if cached.expireAt.Nanosecond() != 0 {
				cache.expires[key] = cached.expireAt
			}

			// TODO: check expires

		} else if err == nil {
			// cache found
			dec := gob.NewDecoder(bytes.NewBuffer(data))
			dec.Decode(&cached)
			c.Writer.WriteHeader(cached.status)
			for k, val := range cached.header {
				for _, v := range val {
					c.Writer.Header().Add(k, v)
				}
			}
			c.Writer.Write(cached.body)
		} else {
			panic(err)
		}
	}
}

func md5String(url string) string {
	h := md5.New()
	io.WriteString(h, url)
	return string(h.Sum(nil))
}

func init() {
	gob.Register(Cached{})
}
