package main

// #cgo LDFLAGS: -lSDL2
/*
#include <SDL2/SDL.h>
#include <SDL2/SDL_audio.h>
#include <string.h>
*/
import "C"
import (
	"os"
	"os/signal"
	"github.com/joao-victor-silva/audio-processor/audio"
)


type ProcessData interface {
	Process(input <- chan byte, output chan <- byte, dataType C.SDL_AudioFormat)
}

type Copy struct {}
type Effect struct {}

func main() {
	sdlManager, err := audio.NewSDL()
	if err != nil {
		panic(err)
	}
	defer sdlManager.Close()

	mic, err := sdlManager.NewAudioDevice(true)
	if err != nil {
		panic("Counldn't open the mic device. " + err.Error())
	}
	defer mic.Close()

	headphone, err := sdlManager.NewAudioDevice(false)
	if err != nil {
		panic("Counldn't open the headphone device" + err.Error())
	}
	defer headphone.Close()


	if mic.AudioFormat() != headphone.AudioFormat() {
		panic("Couldn't use the same audio format for mic and headphones")
	}
	
	mic.Unpause()
	headphone.Unpause()

	copyFromRecord := Copy{}
	go copyFromRecord.Process(mic, headphone, (C.SDL_AudioFormat) (mic.AudioFormat()))

	mainThreadSignals := make(chan os.Signal, 1)
	signal.Notify(mainThreadSignals, os.Interrupt)
	_ = <- mainThreadSignals
}

func (*Copy) Process(inputDevice audio.AudioDevice , outputDevice audio.AudioDevice, audioFormat C.SDL_AudioFormat) {
	for outputDevice.IsChannelOpen() {
		outputDevice.WriteData(inputDevice.ReadData())
	}
}

func (*Effect) Process(inputDevice audio.AudioDevice , outputDevice audio.AudioDevice, audioFormat C.SDL_AudioFormat) {
	for outputDevice.IsChannelOpen() {
		outputDevice.WriteData(inputDevice.ReadData() / 100)
	}
}

// func (*Effect) Process(input <- chan byte, output chan <- byte, audioFormat C.SDL_AudioFormat) {
// 	for true {
// 		if audioFormat == C.AUDIO_F32 {
// 			binaryData := make([]byte, 4)
// 			binaryData[0] = <- input
// 			binaryData[1] = <- input
// 			binaryData[2] = <- input
// 			binaryData[3] = <- input
//
// 			buffer := math.Float32frombits(binary.LittleEndian.Uint32(binaryData))
// 			buffer = buffer / 100
// 			binary.LittleEndian.PutUint32(binaryData, math.Float32bits(buffer))
//
// 			output <- binaryData[0]
// 			output <- binaryData[1]
// 			output <- binaryData[2]
// 			output <- binaryData[3]
// 		}
// 	}
//
// 	// for data := range input {
// 	// }
// }
