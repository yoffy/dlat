package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pSetCommState = kernel32.NewProc("SetCommState")
)

// see also: https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-dcb
type DCB struct {
	DCBlength int32
	BaudRate  int32
	Flags     uint32
	wReserved int16
	XonLim    int16
	XoffLim   int16
	ByteSize  uint8
	Parity    uint8
	StopBits  uint8
	XonChar   int8
	XoffChar  int8
	ErrorChar int8
	EofChar   int8
	EvtChar   int8
}

const (
	DcbFlagsBinary              = 0x00000001
	DcbFlagsParity              = 0x00000002
	DcbFlagsOutxCtsFlow         = 0x00000004
	DcbFlagsOutxDsrFlow         = 0x00000008
	DcbFlagsDtrControlDisable   = 0x00000000
	DcbFlagsDtrControlEnable    = 0x00000010
	DcbFlagsDtrControlHandshake = 0x00000020
	DcbFlagsDsrSensitivity      = 0x00000040
	DcbFlagsTXContinueOnXoff    = 0x00000080
	DcbFlagsOutX                = 0x00000100
	DcbFlagsInX                 = 0x00000200
	DcbFlagsErrorChar           = 0x00000400
	DcbFlagsNull                = 0x00000800
	DcbFlagsRtsControlDisable   = 0x00000000
	DcbFlagsRtsControlEnable    = 0x00001000
	DcbFlagsRtsControlHandshake = 0x00002000
	DcbFlagsAbortOnError        = 0x00004000
)

func SetCommState(handle syscall.Handle, dcb *DCB) error {
	r1, _, e1 := pSetCommState.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(dcb)))
	if r1 == 0 {
		return e1
	}

	return nil
}

type Brightness struct {
	time    time.Time
	isWhite byte
}

type Game struct {
	serialFd    syscall.Handle
	serialChan  chan Brightness
	queuedTime  time.Time
	updatedTime time.Time
	color       uint8
	changing    int
}

func (g *Game) init(comPort string) error {
	fd, err := syscall.Open("\\\\.\\"+comPort, syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}

	dcb := DCB{}
	dcb.DCBlength = int32(unsafe.Sizeof(dcb))
	dcb.BaudRate = 38400
	dcb.Flags =
		DcbFlagsBinary |
			DcbFlagsDtrControlEnable
	dcb.ByteSize = 8
	if err := SetCommState(fd, &dcb); err != nil {
		return err
	}

	g.serialFd = fd
	g.serialChan = make(chan Brightness)
	go g.readSerial(g.serialChan, g.serialFd)

	g.queuedTime = time.Now()

	return nil
}

func (g *Game) finalize() error {
	syscall.Close(g.serialFd)

	return nil
}

func (g *Game) readSerial(ch chan Brightness, fd syscall.Handle) {
	defer syscall.Close(fd)

	buf := make([]byte, 1)
	for {
		if _, err := syscall.Read(fd, buf); err != nil {
			log.Fatalf("%s\n", err)
		}
		ch <- Brightness{time.Now(), buf[0]}
	}
}

func (g *Game) Update() error {
	now := time.Now()

	if g.changing == 1 {
		// now, buffer was swapped
		g.changing = 2
		g.updatedTime = now
		fmt.Printf("%s: -> %d\n", g.updatedTime.Format("15:04:05.000"), g.color)
	}

	// loop for clear g.serialChan
	isLoop := true
	for isLoop {
		select {
		case value := <-g.serialChan:
			if g.changing == 2 {
				diff := value.time.Sub(g.updatedTime)
				fmt.Printf("%s: <- %d (%s)\n", value.time.Format("15:04:05.000"), value.isWhite, diff)
				g.changing = 0
			} else {
				fmt.Printf("%s: <- %d (?)\n", value.time.Format("15:04:05.000"), value.isWhite)
			}

		default:
			isLoop = false
		}
	}

	if now.Second() != g.queuedTime.Second() {
		// queue to change color
		g.queuedTime = now
		mono := uint8(now.Second() % 2)
		g.color = mono
		g.changing = 1
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	c := uint8(g.color * 255)
	screen.Fill(color.RGBA{c, c, c, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 180
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf(
			"usage: dlat COM-PORT\n" +
				"\n" +
				"COM-PORT\n" +
				"\tCOM port name (ex. com5)\n")
		os.Exit(1)
	}

	game := &Game{}
	if err := game.init(os.Args[1]); err != nil {
		log.Fatalf("%s\n", err)
	}
	defer game.finalize()

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("dlat")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
