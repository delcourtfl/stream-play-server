package main

// for now 

// //https://developer.mozilla.org/en-US/docs/Web/API/Screen_Capture_API/Using_Screen_Capture

// import (
// 	"fmt"

// 	"github.com/pion/mediadevices"

// 	"github.com/pion/mediadevices/pkg/driver"

// 	// "github.com/pion/mediadevices/examples/internal/signal"
// 	// "github.com/pion/mediadevices/pkg/frame"
// 	// "github.com/pion/mediadevices/pkg/prop"
// 	// "github.com/pion/webrtc/v3"

// 	// If you don't like x264, you can also use vpx by importing as below
// 	// "github.com/pion/mediadevices/pkg/codec/vpx" // This is required to use VP8/VP9 video encoder
// 	// or you can also use openh264 for alternative h264 implementation
// 	// "github.com/pion/mediadevices/pkg/codec/openh264"
// 	// or if you use a raspberry pi like, you can use mmal for using its hardware encoder
// 	// "github.com/pion/mediadevices/pkg/codec/mmal"
// 	// "github.com/pion/mediadevices/pkg/codec/opus" // This is required to use opus audio encoder
// 	// "github.com/pion/mediadevices/pkg/codec/x264" // This is required to use h264 video encoder

// 	// Note: If you don't have a camera or microphone or your adapters are not supported,
// 	//       you can always swap your adapters with our dummy adapters below.
// 	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
// 	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
// 	// _ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
// 	// _ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
// )

// func main() {

// 	s, err := mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(s)

// 	m := mediadevices.EnumerateDevices()

// 	fmt.Println(m)

// 	typeFilter := driver.FilterVideoRecorder()
// 	fmt.Printf("Type Filter: %#v\n", typeFilter)

// 	notScreenFilter := driver.FilterNot(driver.FilterDeviceType(driver.Screen))
// 	fmt.Printf("Not Screen Filter: %#v\n", notScreenFilter)

// 	filter := driver.FilterAnd(typeFilter, notScreenFilter)
// 	fmt.Printf("Combined Filter: %#v\n", filter)

// }