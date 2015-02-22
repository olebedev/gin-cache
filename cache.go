package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const KEY_PREFIX = "gin:cache:"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type cached struct {
	Status   int
	Body     []byte
	Header   http.Header
	ExpireAt time.Time
}

type Store interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
	Remove(string) error
	Update(string, []byte) error
	Keys() []string
}

type Options struct {
	Store         Store
	Expire        time.Duration
	Headers       []string
	DoNotUseAbort bool
}

func (o *Options) init() {
	if o.Headers == nil {
		o.Headers = []string{
			"User-Agent",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cookie",
			"User-Agent",
		}
	}
}

type Cache struct {
	Store
	options Options
	expires map[string]time.Time
}

type wrappedWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (rw *wrappedWriter) Write(body []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(body)
	if err == nil {
		rw.body.Write(body)
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

		var cch cached
		tohash := c.Request.URL.RequestURI()
		for _, k := range cache.options.Headers {
			if v, ok := c.Request.Header[k]; ok {
				tohash += k
				tohash += strings.Join(v, "")
			}
		}

		key := KEY_PREFIX + md5String(tohash)

		if data, err := cache.Get(key); err == ErrNotFound {
			// cache miss
			writer := c.Writer
			rw := wrappedWriter{ResponseWriter: c.Writer}
			c.Writer = &rw
			c.Next()
			c.Writer = writer

			cch = cached{
				Status: rw.Status(),
				Body:   rw.body.Bytes(),
				Header: http.Header(rw.Header()),
				ExpireAt: func() time.Time {
					if cache.options.Expire == 0 {
						return time.Time{}
					} else {
						return time.Now().Add(cache.options.Expire)
					}
				}(),
			}

			var b bytes.Buffer
			enc := gob.NewEncoder(&b)

			panicIf(enc.Encode(cch))
			cache.Set(key, b.Bytes())

			if cch.ExpireAt.Nanosecond() != 0 {
				cache.expires[key] = cch.ExpireAt
			}

			// TODO: check expires

		} else if err == nil {
			// cache found
			dec := gob.NewDecoder(bytes.NewBuffer(data))
			dec.Decode(&cch)
			c.Writer.WriteHeader(cch.Status)
			for k, val := range cch.Header {
				for _, v := range val {
					c.Writer.Header().Add(k, v)
				}
			}
			c.Writer.Header().Add("X-Gin-Cache", "yes")
			c.Writer.Write(cch.Body)

			if !cache.options.DoNotUseAbort {
				c.Abort()
			}

		} else {
			panic(err)
		}
	}
}

func md5String(url string) string {
	h := md5.New()
	io.WriteString(h, url)
	return hex.EncodeToString(h.Sum(nil))
}

func init() {
	gob.Register(cached{})
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
