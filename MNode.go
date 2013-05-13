package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"net"
)

type MandleSegment struct {
	BotLeft, TopRight complex128
	SizeX, SizeY int
	StartY, EndY int
	Num int
}

type MandleReduceInfo struct {
	Arr []byte
}

func calcMandlebrot(c complex128) int {
	var z complex128
	iterations := 0
	for ;iterations < 256 && real(z) <= 4; iterations++ {
		z = c + (z * z)
	}
	return iterations
}

func (mi *MandleSegment) calcMandleSegment() []byte {
	rln := (mi.EndY - mi.StartY) * mi.SizeX
	fmt.Printf("len = %d\n", rln)
	ret := make([]byte, rln)
	i := 0
	for y := mi.StartY; y < mi.EndY; y++ {
		for x := 0; x < mi.SizeX; x++ {
			a := real(mi.BotLeft) + (float64(x) * (real(mi.TopRight) - real(mi.BotLeft)) / float64(mi.SizeX - 1))
			b := imag(mi.TopRight) - (float64(y) * (imag(mi.TopRight) - imag(mi.BotLeft)) / float64(mi.SizeY - 1))
			c := complex(a, b)
			ret[i] = byte(calcMandlebrot(c))
			i++
		}
	}
	return ret
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify a hostname for the job server!")
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	con, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	var ms MandleSegment
	dec := gob.NewDecoder(con)
	err = dec.Decode(&ms)
	if err != nil {
		panic(err)
	}
	fmt.Println(ms)
	rval := ms.calcMandleSegment()
	rmi := MandleReduceInfo{}
	rmi.Arr = rval
	enc := gob.NewEncoder(con)
	enc.Encode(rmi)

}
