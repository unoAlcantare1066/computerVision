package gopusenc

import (
	"fmt"
	"io"
	"sync"
	"unsafe"
)

/*
#include <stdlib.h>
#include <opusenc.h>
#cgo CFLAGS: -I /usr/include/opus
#cgo LDFLAGS: -lopusenc -lopus

OggOpusEnc *ope_encoder_create_callbacks(const OpusEncCallbacks *callbacks, void *user_data,
    OggOpusComments *comments, opus_int32 rate, int channels, int family, int *error);
#include "opus_callbacks.h"
*/
import "C"

type Family int

type writersCache struct {
	sync.Mutex
	lastId  int
	writers map[int]io.Writer
}

var writers = &writersCache{lastId: 0, writers: make(map[int]io.Writer)}

func (c *writersCache) addWriter(w io.Writer) int {
	c.Lock()
	defer c.Unlock()
	id := c.lastId
	c.writers[id] = w
	c.lastId++

	return id
}

func (c *writersCache) getWriter(id int) (writer io.Writer, found bool) {
	c.Lock()
	defer c.Unlock()
	writer, found = c.writers[id]
	return
}

func (c *writersCache) removeWriter(id int) {
	c.Lock()
	defer c.Unlock()

	delete(c.writers, id)
}

const (
	MonoStereo Family = 0
	Surround   Family = 1
)

type Encoder struct {
	rate       int
	nChannels  int
	familyType Family
	writer     io.Writer
	writerId   int
	comments   *C.OggOpusComments
	enc        *C.struct_OggOpusEnc
	callbacks  C.OpusEncCallbacks
}

func getOpeErrorString(errCode C.int) string {
	cerr := C.ope_strerror(errCode)
	return C.GoString(cerr)
}

func NewEncoder(rate int, nChannels int, family Family, writer io.Writer) *Encoder {
	return &Encoder{rate, nChannels, family, writer, -1, nil, nil,
		C.OpusEncCallbacks{}}
}

//export goWriteCallback
func goWriteCallback(userData unsafe.Pointer, data *C.uchar, len C.int) C.int {
	writerId := *(*int)(userData)
	w, f := writers.getWriter(writerId)
	if !f {
		panic(fmt.Errorf("unknown writer with id"))
	}

	buf := C.GoBytes(unsafe.Pointer(data), len)
	if _, e := w.Write(buf); e != nil {
		return 1
	}

	return 0
}

//export goCloseCallback
func goCloseCallback(userData unsafe.Pointer) C.int {
	writerId := *(*int)(userData)
	writers.removeWriter(writerId)
	return 0
}

func (enc *Encoder) Init() error {
	var err C.int = 0

	enc.comments = C.ope_comments_create()
	enc.callbacks.write = (C.ope_write_func)(C.CallWriteCb)
	enc.callbacks.close = (C.ope_close_func)(C.CallCloseCb)

	enc.writerId = writers.addWriter(enc.writer)

	enc.enc = C.ope_encoder_create_callbacks(&enc.callbacks, unsafe.Pointer(&enc.writerId), enc.comments,
		C.int(enc.rate), C.int(enc.nChannels), C.int(enc.familyType), &err)
	if err != 0 {
		writers.removeWriter(enc.writerId)
		return fmt.Errorf("error initializing opus encoder: %s", getOpeErrorString(err))
	}

	return nil
}

func (enc *Encoder) Finish() {
	enc.Drain()
	C.free(unsafe.Pointer(enc.comments))
	writers.removeWriter(enc.writerId)

	C.ope_encoder_destroy(enc.enc)
}

func (enc *Encoder) Encode(nSamplesPerChannel int, pcm []int16) error {
	var err C.int = 0

	err = C.ope_encoder_write(enc.enc, (*C.opus_int16)(&pcm[0]), C.int(nSamplesPerChannel))
	if err != 0 {
		return fmt.Errorf("ope_encoder_write() error: %s\n", getOpeErrorString(err))
	}

	return nil
}

func (enc *Encoder) Drain() {
	if enc.enc != nil {
		C.ope_encoder_drain(enc.enc)
	}
}
