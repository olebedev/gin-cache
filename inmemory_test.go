package cache

import (
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	im := InMemory{
		hash: map[string][]byte{
			"key": []byte{1, 2, 3},
		},
	}
	// im.Set("key", []byte{1, 2, 3})
	res, err := im.Get("key")
	expect(t, err, nil)
	expect(t, len(res), 3)

	_, err = im.Get("__key__")
	expect(t, err, ErrNotFound)
}

func TestIMSet(t *testing.T) {
	im := InMemory{hash: map[string][]byte{}}

	err := im.Set("key", []byte{1, 2, 3})
	expect(t, err, nil)

	res, ok := im.hash["key"]
	expect(t, ok, true)
	expect(t, len(res), 3)

	err = im.Set("key", []byte{1, 2, 3})
	expect(t, err, ErrAlreadyExists)
}

func TestIMRemove(t *testing.T) {
	im := InMemory{
		hash: map[string][]byte{
			"key": []byte{1, 2, 3},
		},
	}

	err := im.Remove("key")
	expect(t, err, nil)

	err = im.Remove("__key__")
	expect(t, err, ErrNotFound)

	expect(t, len(im.hash), 0)
}

// update
func TestIMUpdate(t *testing.T) {
	im := InMemory{
		hash: map[string][]byte{
			"key": []byte{1, 2, 3},
		},
	}

	err := im.Update("key", []byte{1})
	expect(t, err, nil)
	expect(t, len(im.hash["key"]), 1)

	err = im.Update("__key__", []byte{1})
	expect(t, err, ErrNotFound)
}

// keys
func TestIMKeys(t *testing.T) {
	im := InMemory{
		hash: map[string][]byte{
			"key": []byte{1, 2, 3},
		},
	}

	keys := im.Keys()
	expect(t, len(keys), 1)
	expect(t, keys[0], "key")

	im.hash["__key__"] = []byte{}

	keys = im.Keys()
	expect(t, len(keys), 2)
	expect(t, keys[0], "key")
	expect(t, keys[1], "__key__")
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
