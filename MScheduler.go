package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"math"
	"net"
	"fmt"
	"encoding/gob"
)

type MandleSegment struct {
	BotLeft, TopRight complex128
	SizeX, SizeY int
	StartY, EndY int
	Num int
}

type MandlebrotInfo struct {
	BotLeft, TopRight complex128
	x,y int
	points [][]byte
	complete int
}

type MandleReduceInfo struct {
	Arr []byte
}

func NewMandlebrot(x, y, segments int) *MandlebrotInfo {
	m := new(MandlebrotInfo)
	m.x = x
	m.y = y
	m.BotLeft = complex(-2.0, -1)
	m.TopRight = complex(1,1)
	m.points = make([][]byte, segments)
	return m
}

func (mi *MandlebrotInfo) Render(Color func(b byte) color.Color, outfile string) error {
	x, y := 0,0
	rect := image.Rect(0,0,mi.x,mi.y)
	img := image.NewRGBA(rect)
	for i := 0; i < len(mi.points); i++ {
		for j := 0; j < len(mi.points[i]); j++ {
			if x == mi.x {
				x = 0
				y++
			}
			img.Set(x,y, Color(mi.points[i][j]))
			x++
		}
	}
	fi, err := os.Create(outfile)
	if err != nil {
		return err
	}
	err = png.Encode(fi, img)
	return err
}


func BlueishGradient(b byte) color.Color {
	col := color.RGBA{}
	col.R = uint8(255 * math.Sin(float64(b)))
	col.G = uint8(255 * (math.Pow(math.Sin(float64(b)), 2) * float64(b)) / 2)
	col.B = uint8(255 * math.Sin(float64(b)))
	col.A = 255
	return col
}

func BlackAndWhite(b byte) color.Color {
	col := color.Gray{}
	if b == 0 {
		col.Y = 0
	} else {
		col.Y = 255
	}
	return col
}

func (mi *MandlebrotInfo) GetNextTask() *MandleSegment {
	if mi.complete == len(mi.points) {
		return nil
	}

	ms := new(MandleSegment)
	ms.SizeX = mi.x
	ms.SizeY = mi.y
	ms.BotLeft = mi.BotLeft
	ms.TopRight = mi.TopRight

	N := 0
	for i := mi.complete; i < len(mi.points); i++ {
		if mi.points[i] == nil {
			N = i
			break
		}
	}
	mi.complete++

	ms.Num = N
	count := len(mi.points)
	ms.StartY = N * (mi.y / count)
	ms.EndY = (N + 1) * (mi.y / count)
	fmt.Printf("%d %d\n", ms.StartY, ms.EndY)
	return ms
}

func handleConnection(c net.Conn, ms *MandleSegment, m *MandlebrotInfo, ch chan bool) {
	enc := gob.NewEncoder(c)
	fmt.Println(ms)
	err := enc.Encode(ms)
	if err != nil {
		fmt.Println(err)
		return
	}
	dec := gob.NewDecoder(c)
	var mri MandleReduceInfo
	err = dec.Decode(&mri)
	if err != nil {
		panic(err)
	}
	m.points[ms.Num] = mri.Arr
	ch <- true
}

func main() {
	m := NewMandlebrot(6000,4000, 8)

	laddr, _ := net.ResolveTCPAddr("tcp", ":8080")
	list, err := net.ListenTCP("tcp", laddr)

	if err != nil {
		panic(err)
	}
	fc := make(chan bool)
	go func() {
		for {
			<-fc
			if m.complete == len(m.points) {
				list.Close()
			}
		}
	}()

	for {
		con, err := list.Accept()
		if err != nil {
			fmt.Println("finished all segments, exiting network loop to start rendering")
			break
		}
		go func() {
			ms := m.GetNextTask()
			if ms == nil {
				return
			}
			handleConnection(con, ms, m, fc)
		}()
	}
	m.Render(BlueishGradient, "nettest.png")
}

