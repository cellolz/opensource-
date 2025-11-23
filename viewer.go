package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"
	"flag"
	"strconv"

	"golang.org/x/sys/windows"
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	gdi32   = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW  = user32.NewProc("RegisterClassExW")
	procCreateWindowExW   = user32.NewProc("CreateWindowExW")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procUpdateWindow      = user32.NewProc("UpdateWindow")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procLoadImageW        = user32.NewProc("LoadImageW")
	procGetDC             = user32.NewProc("GetDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")
	procBitBlt            = gdi32.NewProc("BitBlt")
	procCreateCompatibleDC= gdi32.NewProc("CreateCompatibleDC")
	procSelectObject      = gdi32.NewProc("SelectObject")
	procDeleteDC          = gdi32.NewProc("DeleteDC")
	procDeleteObject      = gdi32.NewProc("DeleteObject")

	zfile = ""
	ztime = "1000"
)

const (
	WS_POPUP      = 0x80000000
	WS_VISIBLE    = 0x10000000
	SW_SHOWMAXIMIZED = 3
	LR_LOADFROMFILE = 0x00000010
	IMAGE_BITMAP   = 0
	SRCCOPY        = 0x00CC0020


    	WM_PAINT = 0x000F
   	WM_CLOSE = 0x0010
    	WM_KEYDOWN = 0x0100

)

type (
	HWND    windows.Handle
	HBITMAP windows.Handle
	HDC     windows.Handle
)

type WNDCLASSEXW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       windows.Handle
}

type MSG struct {
	hwnd    HWND
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

func mustUTF16Ptr(s string) *uint16 {
	ptr, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return ptr
}

func wndProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
    switch msg {
    case WM_CLOSE:
        fmt.Println("Janela principal encerrada...")
        os.Exit(0)
    case WM_KEYDOWN:
        //fmt.Printf("Mensagem: %04X %08X\n", msg, lParam)
        if lParam == 0x3d0001 {
                fmt.Println("Pressionado F3 para encerrar...")
                os.Exit(0)
        }
    default:
        //fmt.Printf("Mensagem: %04X\n", msg)
    }

	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd), uintptr(msg), wParam, lParam,
	)
	return ret
}

func registerWindowClass(className string, hInstance windows.Handle) error {
	var wc WNDCLASSEXW
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.lpfnWndProc = windows.NewCallback(wndProc)
	wc.hInstance = hInstance
	wc.lpszClassName = mustUTF16Ptr(className)

	r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return err
	}
	return nil
}

func createFullscreenWindow(className, title string, hInstance windows.Handle) (HWND, error) {
	hwndRaw, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(mustUTF16Ptr(className))),
		uintptr(unsafe.Pointer(mustUTF16Ptr(title))),
		WS_POPUP|WS_VISIBLE,
		0, 0,
		1920, 	//windows.GetSystemMetrics(0), // largura tela
		1024, 	//windows.GetSystemMetrics(1), // altura tela
		0, 0,
		uintptr(hInstance),
		0,
	)
	if hwndRaw == 0 {
		return 0, err
	}

	procShowWindow.Call(hwndRaw, SW_SHOWMAXIMIZED)
	procUpdateWindow.Call(hwndRaw)
	return HWND(hwndRaw), nil
}

func loadBMPtoHBITMAP(filename string) HBITMAP {
	h, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(mustUTF16Ptr(filename))),
		IMAGE_BITMAP,
		0, 0,
		LR_LOADFROMFILE,
	)
	return HBITMAP(h)
}

func blitBitmap(hwnd HWND, hBitmap HBITMAP) {
	hdcWnd, _, _ := procGetDC.Call(uintptr(hwnd))
	hdcMem, _, _ := procCreateCompatibleDC.Call(hdcWnd)

	procSelectObject.Call(hdcMem, uintptr(hBitmap))

	width := 1920	//windows.GetSystemMetrics(0)
	height := 1080	//windows.GetSystemMetrics(1)

	procBitBlt.Call(hdcWnd, 0, 0, uintptr(width), uintptr(height),
		hdcMem, 0, 0, SRCCOPY)

	procDeleteDC.Call(hdcMem)
	procReleaseDC.Call(uintptr(hwnd), hdcWnd)
}

func main() {
	fmt.Println("Iniciando processo...")
	fmt.Println("Para interromper o programa pressione <F3>!")


	flag.StringVar(&zfile, "f", "zcode4_%d.bmp", "Nome do arquivo a ser processado.")
	flag.StringVar(&ztime, "t", "1000", "Intervalo para atualizar em ms (default: 1000 = 1s)")

	flag.Parse()

	zntime,_ := strconv.Atoi(ztime)

	fmt.Println("Filename:", zfile, ", Refresh interval:", zntime)



	hInstance := windows.CurrentProcess()

	className := "MyFullscreenClass"
	if err := registerWindowClass(className, hInstance); err != nil {
		log.Fatal(err)
	}

	hwnd, err := createFullscreenWindow(className, "BMP Viewer", hInstance)
	if err != nil {
		log.Fatal(err)
	}

	seq := 0
	go func() {
		for {
			filename := fmt.Sprintf(zfile, seq)
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				seq = 1
				continue
			}

			hBitmap := loadBMPtoHBITMAP(filename)
			if hBitmap != 0 {
				// delay entre bitmap...
				time.Sleep(time.Duration(zntime) * time.Millisecond)

				fmt.Println("filename:", filename)


				blitBitmap(hwnd, hBitmap)
				procDeleteObject.Call(uintptr(hBitmap))
			} else {
				log.Println("Falha ao carregar:", filename)
			}

			seq++
		}
	}()

	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
