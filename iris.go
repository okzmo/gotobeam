package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
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

type Atoms struct {
	inner []string
}

func (a *Atoms) findId(needle string) uint {
	for i, name := range a.inner {
		if name == needle {
			return uint(i + 1)
		}
	}
	a.inner = append(a.inner, needle)
	return uint(len(a.inner))
}

type Tag uint8

const (
	U Tag = 0
	I     = 1
	A     = 2
	X     = 3
	Y     = 4
	F     = 5
	H     = 6
	Z     = 7
)

type OpCode uint8

const (
	Label      OpCode = 1
	FuncInfo          = 2
	IntCodeEnd        = 3
	Return            = 19
	Move              = 64
)

func encode(tag Tag, n int32) []byte {
	var result []byte

	if n < 0 {
		// negative number
	} else if n < 16 {
		var tag = uint8(tag)
		var n = uint8(n)
		result = []byte{n<<4 | tag}
	} else if n < 0x800 {
		var tag = uint32(tag)
		var n = uint32(n)
		var firstByte = uint8(((n >> 3) & uint32(0b11100000)) | tag | uint32(0b00001000))
		var secondByte = uint8(n & 0xFF)
		result = []byte{firstByte, secondByte}
	} else {
		// large number
	}

	return result
}

func paddingChunk(chunk []byte) []byte {
	pad := len(chunk) % 4
	if pad != 0 {
		chunk = append(chunk, bytes.Repeat([]byte{0}, 4-pad)...)
	}

	return chunk
}

func codeChunk(program map[string]uint, atoms *Atoms, labels map[uint32]uint32) []byte {
	var label_count uint32 = 0
	var function_count uint32 = 0
	var code []byte

	for name, v := range program {
		function_count++

		label_count++
		code = append(code, uint8(Label))
		code = append(code, encode(Tag(U), int32(label_count))...)

		code = append(code, uint8(FuncInfo))
		code = append(code, encode(Tag(A), int32(atoms.findId("iris")))...) // module name
		id := atoms.findId(name)
		code = append(code, encode(Tag(A), int32(id))...) // function name
		code = append(code, encode(Tag(U), 0)...)         // arity

		label_count++
		code = append(code, uint8(Label))
		code = append(code, encode(Tag(U), int32(label_count))...)
		labels[uint32(id)] = label_count

		code = append(code, uint8(Move))
		code = append(code, encode(Tag(I), int32(v))...)
		code = append(code, encode(Tag(X), 0)...)
		code = append(code, uint8(Return))
	}
	code = append(code, uint8(IntCodeEnd))
	label_count++

	var sub_size uint32 = 16
	var instruction_set uint32 = 0
	var opcode_max uint32 = 169

	var chunk []byte
	chunk = append(chunk, toBeBytes(sub_size)...)
	chunk = append(chunk, toBeBytes(instruction_set)...)
	chunk = append(chunk, toBeBytes(opcode_max)...)
	chunk = append(chunk, toBeBytes(label_count)...)
	chunk = append(chunk, toBeBytes(function_count)...)
	chunk = append(chunk, code...)

	var fullChunk []byte
	fullChunk = append(fullChunk, []byte("Code")...)
	fullChunk = append(fullChunk, toBeBytes(uint32(len(chunk)))...)
	chunk = paddingChunk(chunk)
	fullChunk = append(fullChunk, chunk...)

	return fullChunk
}

func atomChunk(atoms *Atoms) []byte {
	var chunk []byte
	chunk = append(chunk, toBeBytes(uint32(len(atoms.inner)))...)
	for _, atom := range atoms.inner {
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

func importsChunk() []byte {
	var import_count uint32 = 0

	var chunk []byte
	chunk = append(chunk, toBeBytes(import_count)...)

	var fullChunk []byte
	fullChunk = append(fullChunk, []byte("ImpT")...)
	fullChunk = append(fullChunk, toBeBytes(uint32(len(chunk)))...)
	chunk = paddingChunk(chunk)
	fullChunk = append(fullChunk, chunk...)

	return fullChunk
}

func exportsChunk(labels map[uint32]uint32) []byte {
	var export_count uint32 = uint32(len(labels))

	var chunk []byte
	chunk = append(chunk, toBeBytes(export_count)...)

	for id, label := range labels {
		chunk = append(chunk, toBeBytes(id)...)        // Function name
		chunk = append(chunk, toBeBytes(uint32(0))...) // Arity
		chunk = append(chunk, toBeBytes(label)...)     // Label
	}

	var fullChunk []byte
	fullChunk = append(fullChunk, []byte("ExpT")...)
	fullChunk = append(fullChunk, toBeBytes(uint32(len(chunk)))...)
	chunk = paddingChunk(chunk)
	fullChunk = append(fullChunk, chunk...)

	return fullChunk
}

func stringChunk() []byte {
	var fullChunk []byte
	fullChunk = append(fullChunk, []byte("StrT")...)
	fullChunk = append(fullChunk, toBeBytes(uint32(0))...)

	return fullChunk
}

func main() {
	f, err := os.Open("test.iris")
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	program := make(map[string]uint)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			splitted := strings.Split(strings.ReplaceAll(line, " ", ""), "=")
			res, _ := strconv.ParseUint(splitted[1], 10, 64)
			fmt.Println(res)
			program[splitted[0]] = uint(res)
		}
	}

	var atomsEntries Atoms
	var labels = make(map[uint32]uint32)

	code := codeChunk(program, &atomsEntries, labels)
	imports := importsChunk()
	exports := exportsChunk(labels)
	string := stringChunk()
	atoms := atomChunk(&atomsEntries)

	beam := make([]byte, 0)
	beam = append(beam, []byte("BEAM")...)
	beam = append(beam, code...)
	beam = append(beam, imports...)
	beam = append(beam, exports...)
	beam = append(beam, string...)
	beam = append(beam, atoms...)

	var bytes []byte
	bytes = append(bytes, []byte("FOR1")...)
	bytes = append(bytes, toBeBytes(uint32(len(beam)))...)
	bytes = append(bytes, beam...)

	f, err = os.Create("iris.beam")
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	defer f.Close()

	f.Write(bytes)
}
