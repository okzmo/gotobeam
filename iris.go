package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type Uint32 uint32
type Uint8 uint8
type ToBigEndianBytes interface {
	ToBeBytes() []byte
}

func toBeBytes(n interface{}) []byte {
	switch t := n.(type) {
	case uint32:
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, t)
		return bytes
	case uint8:
		return []byte{t}
	default:
		return nil
	}
}

func paddingChunk(chunk []byte) []byte {
	pad := len(chunk) % 4
	fmt.Println(chunk, pad)
	if pad != 0 {
		chunk = append(chunk, bytes.Repeat([]byte{0}, 4-pad)...)
	}

	return chunk
}

func codeChunk() []byte {
	var sub_size uint32 = 16
	var instruction_set uint32 = 0
	var opcode_max uint32 = 169
	var label_count uint32 = 0
	var function_count uint32 = 0

	var chunk []byte
	chunk = append(chunk, toBeBytes(sub_size)...)
	chunk = append(chunk, toBeBytes(instruction_set)...)
	chunk = append(chunk, toBeBytes(opcode_max)...)
	chunk = append(chunk, toBeBytes(label_count)...)
	chunk = append(chunk, toBeBytes(function_count)...)

	var fullChunk []byte
	fullChunk = append(fullChunk, []byte("Code")...)
	fullChunk = append(fullChunk, toBeBytes(uint32(len(chunk)))...)
	chunk = paddingChunk(chunk)
	fullChunk = append(fullChunk, chunk...)

	return fullChunk
}

func atomChunk(atoms []string) []byte {
	var chunk []byte
	chunk = append(chunk, toBeBytes(uint32(len(atoms)))...)
	for _, atom := range atoms {
		chunk = append(chunk, toBeBytes(uint8(len(atom)))...)
		chunk = append(chunk, []byte(atom)...)
	}

	var fullChunk []byte
	fullChunk = append(fullChunk, []byte("AtU8")...)
	fullChunk = append(fullChunk, toBeBytes(uint32(len(chunk)))...)
	chunk = paddingChunk(chunk)
	fullChunk = append(fullChunk, chunk...)

	return fullChunk
}

func main() {
	f, err := os.Create("iris.beam")
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	defer f.Close()

	code := codeChunk()
	atoms := atomChunk([]string{"iris", "hello", "world"})

	beam := make([]byte, 0)
	beam = append(beam, []byte("BEAM")...)
	beam = append(beam, code...)
	beam = append(beam, atoms...)

	var bytes []byte
	bytes = append(bytes, []byte("FOR1")...)
	bytes = append(bytes, toBeBytes(uint32(len(beam)))...)
	bytes = append(bytes, beam...)

	f.Write(bytes)
}
