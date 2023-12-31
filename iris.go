package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func codeChunk() []byte {

	sub_size := 16
	instruction_set := 0
	opcode_max := 169
	label_count := 0
	function_count := 0
	chunk := make([]byte, 0)
	chunk = append(chunk, byte(sub_size))
	chunk = append(chunk, byte(instruction_set))
	chunk = append(chunk, byte(opcode_max))
	chunk = append(chunk, byte(label_count))
	chunk = append(chunk, byte(function_count))

	return []byte{}
}

func main() {
	f, err := os.Create("iris.beam")
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	defer f.Close()

	code := codeChunk()

	beam := make([]byte, 0)
	beam = append(beam, []byte("BEAM")...)
	beam = append(beam, code...)

	bytes := []byte{}
	bytes = append(bytes, []byte("FOR1")...)
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, uint32(len(beam)))
	bytes = append(bytes, temp...)

	f.Write(bytes)
}
