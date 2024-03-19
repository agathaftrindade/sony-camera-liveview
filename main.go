package main

import (
	"encoding/binary"
	"io"
	"log"
	liveview "sony_camera_liveview/src"
	"strconv"
	"sync"
	"time"

	"github.com/edwingeng/deque/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const TARGET_FPS = 16

// func StreamLiveview(camera *liveview.SonyCamera, out chan []byte) {
func StreamLiveview(camera *liveview.SonyCamera, out *deque.Deque[[]byte], lock *sync.Mutex) {
	packetsChan := make(chan *liveview.Packet, 300)

	camera.StartStreaming(TARGET_FPS, packetsChan)
	for packet := range packetsChan {
		if packet.IsFrameData() {
			// out <- packet.JpgData
			lock.Lock()
			out.PushFront(packet.JpgData)
			lock.Unlock()
		}
	}
}

func PublishToRTSP(in io.Reader, rtspServerURL string) {
	// ffmpeg -re -f 'image2pipe' -i - -c:v 'libx264' -bufsize '64M' -maxrate '4M' -rtsp_transport 'tcp' -f 'rtsp' rtsp://localhost:8554/live.stream
	err := ffmpeg.
		Input("pipe:", ffmpeg.KwArgs{
			"re":        "",
			"f":         "image2pipe",
			"framerate": strconv.Itoa(TARGET_FPS),
		}).
		Output(rtspServerURL, ffmpeg.KwArgs{
			"c:v":            "libx264",
			"preset":         "veryfast",
			"tune":           "zerolatency",
			"f":              "rtsp",
			"rtsp_transport": "tcp",
			"pix_fmt":        "yuv420p",
			// "b":              "100k",
			// "r": strconv.Itoa(TARGET_FPS),

			// "q":              "1",
			// "bufsize":        "64M",
			// "maxrate":        "4M",
		}).
		ErrorToStdOut().
		WithInput(in).
		Run()

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	serverURL := "rtsp://192.168.15.11:8554/live.stream"

	camera := liveview.NewSonyCamera("http://192.168.122.1:10000/sony")
	err := camera.StartLiveView()
	if err != nil {
		panic(err)
	}

	pipeOut, pipeIn := io.Pipe()
	// jpgChan := make(chan []byte, 30*10)
	dq := deque.NewDeque[[]byte]()
	var dqLock sync.Mutex

	go StreamLiveview(&camera, dq, &dqLock)
	go PublishToRTSP(pipeOut, serverURL)

	now := time.Now()
	then := now

	msPerFrame := 1000.0 / TARGET_FPS

	for {

		dqLock.Lock()
		// frame, ok := dq.TryPopFront()
		frame, ok := dq.TryPopBack()
		dqLock.Unlock()

		if !ok {
			continue
		}
		binary.Write(pipeIn, binary.BigEndian, frame)

		now = time.Now()
		elapsed := now.Sub(then)

		toSleep := int64(msPerFrame) - elapsed.Milliseconds()
		if toSleep > 0 {
			// fmt.Printf("Sleeping for %dms\n", toSleep)
			time.Sleep(time.Duration(toSleep) * time.Millisecond)
			dqLock.Lock()
			dq.Clear()
			dqLock.Unlock()
			continue
		}

		// log.Printf("Writing frame %d", packet.CommonHeader.SequenceNumber)

		then = now
		// fmt.Printf("\rFrametime: %f s (%f fps) (target: %d)", elapsed.Seconds(), 1.0/elapsed.Seconds(), TARGET_FPS)
	}

}
