package tinybtree

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

type fieldValuesSlot int

type itemT struct {
	id              string
	obj             geojson.Object
	fieldValuesSlot fieldValuesSlot
}

func itemSaver(w io.Writer, value interface{}) (err error) {
	item := value.(itemT)
	if err = saveString(w, item.id); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, uint64(item.fieldValuesSlot)); err != nil {
		return
	}
	if err = saveString(w, item.obj.JSON()); err != nil {
		return
	}
	return
}

var parseOpts = &geojson.ParseOptions{}

func itemLoader(r io.Reader) (value interface{}, err error) {
	var item itemT
	buf := make([]byte, 0)
	if item.id, buf, err = loadString(r, buf); err != nil {
		return
	}
	var word uint64
	if err = binary.Read(r, binary.BigEndian, &word); err != nil {
		return
	}
	item.fieldValuesSlot = fieldValuesSlot(word)
	var jsonString string
	if jsonString, buf, err = loadString(r, buf); err != nil {
		return
	}
	if item.obj, err = geojson.Parse(jsonString, parseOpts); err != nil {
		return
	}
	return item, nil
}

func TestSaveLoadBTree256(t *testing.T) {
	var tr BTree
	var n int

	for _, i := range rand.Perm(256) {
		key := fmt.Sprintf("key%d", i)
		tr.Set(
			key,
			itemT{
				id: key,
				obj: geojson.NewPoint(
					geometry.Point{X: float64(i)/10, Y: float64(i)/10}),
				fieldValuesSlot: fieldValuesSlot(i),
			})
		n++
		if tr.Len() != n {
			t.Fatalf("expected 256, got %d", n)
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
		oi := ov.(itemT)
		ni := nv.(itemT)
		if oi.id != ni.id {
			t.Fatal("expected equal")
		}
		if oi.fieldValuesSlot != ni.fieldValuesSlot {
			t.Fatal("expected equal")
		}
		if oi.obj.JSON() != ni.obj.JSON() {
			t.Fatal("expected equal")
		}
	}

}
