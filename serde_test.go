package tinybtree

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
)

func itemSaver(w io.Writer, value interface{}) (err error) {
	item := value.(uint32)
	if err = binary.Write(w, binary.BigEndian, item); err != nil {
		return
	}
	return
}

func itemLoader(r io.Reader, obuf []byte) (value interface{}, buf []byte, err error) {
	buf = obuf[:]
	var item uint32
	if err = binary.Read(r, binary.BigEndian, &item); err != nil {
		return
	}
	return item, buf,nil
}

func TestSaveLoadBTree256(t *testing.T) {
	var tr BTree
	var n int

	for _, i := range rand.Perm(256) {
		tr.Set(fmt.Sprintf("key%d", i), uint32(i))
		n++
		if tr.Len() != n {
			t.Fatalf("expected %d, got %d", n, tr.Len())
		}
	}
	var f *os.File
	var err error
	fileName := "/tmp/tree_save"
	f, err = os.Create(fileName)
	if err != nil {
		t.Fatal("creating failed")
	}

	if err = tr.Save(f, itemSaver); err != nil {
		t.Fatal("saving failed")
	}
	if f.Close() != nil {
		t.Fatal("closing failed")
	}

	f, err = os.Open(fileName)
	if err != nil {
		t.Fatal("opening failed")
	}

	var newTr BTree
	if newTr, err = Load(f, itemLoader); err != nil {
		t.Fatal("loading failed")
	}

	if f.Close() != nil {
		t.Fatal("closing failed")
	}

	fmt.Printf("Orig tree: height %d length %d\n", tr.height, tr.length)
	fmt.Printf("New tree: height %d length %d\n", newTr.height, newTr.length)

	fmt.Printf("Old tree: %v\n", tr.root)
	fmt.Printf("New tree: %v\n", newTr.root)

	for _, i := range rand.Perm(256) {
		key := fmt.Sprintf("key%d", i)
		ov, _ := tr.Get(key)
		nv, ok := newTr.Get(key)
		if !ok {
			t.Fatal("expected true")
		}
		if ov.(uint32) != nv.(uint32) {
			t.Fatal("expected equal")
		}
	}
}
