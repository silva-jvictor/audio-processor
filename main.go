package main

// #cgo LDFLAGS: -lSDL2
/*
#include <SDL2/SDL.h>
#include <SDL2/SDL_audio.h>
#include <string.h>
extern void fillBuffer(void *userdata, Uint8 * stream, int len);
extern void readBuffer(void *userdata, Uint8 * stream, int len);
static SDL_AudioCallback get_fn_writeptr() {
    return fillBuffer;
}
static SDL_AudioCallback get_fn_readptr() {
    return readBuffer;
}
*/
import "C"
import (
	"flag"
	"fmt"
	"os"
	"unsafe"
)

type UserData struct {
	data []byte
	pos  int
}

func main() {
	ret := C.SDL_Init(C.SDL_INIT_AUDIO)
	if ret < 0 {
		_ = fmt.Errorf("Error")
		os.Exit(1)
	}
	defer C.SDL_Quit()
	defer func() {
		err := C.GoString(C.SDL_GetError())
		if err != "" {
			fmt.Println("SDL error: ", err)
		} else {
			fmt.Println("Exiting without errors")
		}
	}()

	var isCapture bool
	flag.BoolVar(&isCapture, "input", false, "Set device type to input")
	toCInt := map[bool]C.int{
		true:  C.int(1),
		false: C.int(0),
	}

	flag.Parse()

	nInputDevices := C.SDL_GetNumAudioDevices(toCInt[isCapture])
	for i := C.int(0); i < nInputDevices; i++ {
		deviceName := C.GoString(C.SDL_GetAudioDeviceName(i, toCInt[isCapture]))
		fmt.Println("Device", i, ":", deviceName)
	}

	if nInputDevices == 0 {
		panic("No input devices found")
	}

	var deviceName *C.char
	if isCapture {
		deviceName = C.SDL_GetAudioDeviceName(0, toCInt[isCapture])
	} else {
		deviceName = C.SDL_GetAudioDeviceName(1, toCInt[isCapture])
	}

	var desired, obtained C.SDL_AudioSpec
	var desiredPointer = unsafe.Pointer(&desired)

	C.SDL_memset(desiredPointer, 0, C.sizeof_SDL_AudioSpec)

	dataWanted := 96000 * 4

	var userdata UserData
	userdata.data = make([]byte, dataWanted)
	userdata.pos = 0
	dataPointer := (uintptr)(unsafe.Pointer(&userdata)) ^ 0xFFFFFFFF

	desired.freq = 48000
	desired.format = C.AUDIO_F32
	desired.channels = 1
	desired.samples = 2048

	desired.userdata = (unsafe.Pointer)(dataPointer)

	if isCapture {
		desired.callback = C.get_fn_writeptr()
	} else {
		desired.callback = C.get_fn_readptr()
	}

	deviceId := C.SDL_OpenAudioDevice(deviceName, toCInt[isCapture], &desired, &obtained, C.SDL_AUDIO_ALLOW_ANY_CHANGE)

	if deviceId == 0 {
		panic("Counldn't open device")
	}
	defer C.SDL_CloseAudioDevice(deviceId)

	if isCapture {
		C.SDL_PauseAudioDevice(deviceId, toCInt[false])
		C.SDL_Delay(2020)
		err := os.WriteFile("data.bin", userdata.data, 0644)
		if err != nil {
			panic("Couldn't write file")
		}
	} else {
		dataFromFile, err := os.ReadFile("data.bin")
		if err != nil {
			panic("Couldn't read file")
		}
		copy(userdata.data, dataFromFile)
		C.SDL_PauseAudioDevice(deviceId, toCInt[false])
		C.SDL_Delay(2020)

	}
	C.SDL_PauseAudioDevice(deviceId, toCInt[true])
}

//export fillBuffer
func fillBuffer(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	userdata = unsafe.Pointer(uintptr(userdata) ^ 0xFFFFFFFF)
	userdataPointer := (*UserData)(userdata)

	var size int
	size = len(userdataPointer.data) - userdataPointer.pos

	if size >= int(length) {
		copy(userdataPointer.data[userdataPointer.pos:], C.GoBytes(unsafe.Pointer(stream), length))
		userdataPointer.pos += int(length)
	} else {
		copy(userdataPointer.data[userdataPointer.pos:], C.GoBytes(unsafe.Pointer(stream), C.int (size)))
		userdataPointer.pos = len(userdataPointer.data)
	}
}

//export readBuffer
func readBuffer(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	userdata = unsafe.Pointer(uintptr(userdata) ^ 0xFFFFFFFF)
	userdataPointer := (*UserData)(userdata)

	var size int
	size = len(userdataPointer.data) - userdataPointer.pos

	if size >= int(length) {
		C.memcpy(unsafe.Pointer(stream), unsafe.Pointer(&userdataPointer.data[userdataPointer.pos]), C.ulong(length))
		userdataPointer.pos += int(length)
	} else if size != 0 {
		C.memcpy(unsafe.Pointer(stream), unsafe.Pointer(&userdataPointer.data[userdataPointer.pos]), C.ulong(size))
		userdataPointer.pos = len(userdataPointer.data)
		C.memset(unsafe.Pointer((uintptr(unsafe.Pointer(stream)))+uintptr(size)), 0x00, C.ulong(int(length)-size))
	} else {
		C.memset(unsafe.Pointer(stream), 0x00, C.ulong(length))
	}
}
