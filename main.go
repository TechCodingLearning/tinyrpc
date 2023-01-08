package tinyrpc // Package tinyrpc main

import (
	"encoding/binary"
	"fmt"
)

func PutUvarint(buf []byte, x uint64) int {
	i := 0
	fmt.Printf("byte: %v\n", byte(x))
	for x >= 0x80 {
		buf[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	buf[i] = byte(x)
	return i + 1
}

func main() {
	//buf := make([]byte, 100)
	//var x uint64 = 260
	//fmt.Printf("%x\n", x)
	//size := PutUvarint(buf, x)
	//fmt.Printf("%x\n", buf[:size])
	//s := "abcdefghij"
	//str := strings.Repeat(s, 25) + "aaaaaaaaaa"
	//data := make([]byte, 300)
	//size := header.WriteString(data, str)
	//fmt.Println(size)
	//fmt.Printf("%x\n", data[0])
	//ss, n := header.ReadString(data)
	//fmt.Println(ss)
	//fmt.Println(n)
	//
	var x int64 = 127
	fmt.Println("%v\n", x)
	fmt.Printf("%x\n", x)
	buf := make([]byte, binary.MaxVarintLen64)
	y := binary.PutVarint(buf, x)
	fmt.Println(buf[:y])
	fmt.Printf("%x\n", buf[:y])
}
