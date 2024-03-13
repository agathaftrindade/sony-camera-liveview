package main

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"sony_camera_liveview/src/liveview"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

/*
rtsp_server = 'rtsp://localhost:31415/live.stream'
command = [ffmpeg,
           '-re', # Read at native framerate

           '-f', 'rawvideo',  # Apply raw video as input - it's more efficient than encoding each frame to PNG
           '-s', f'{img_width}x{img_height}',
           '-pixel_format', 'bgr24',
           '-r', f'{fps}',
           '-i', '-',

           '-pix_fmt', 'yuv420p',
           '-c:v', 'libx264',
           '-bufsize', '64M',
           '-maxrate', '4M',

           '-rtsp_transport', 'tcp',
           '-f', 'rtsp',
           rtsp://localhost:31415/live.stream
		   ]
ffmpeg -re -f 'rawvideo' -s '{640}x{360}' -pixel_format 'bgr24' -r 30 -i - -pix_fmt 'yuv420p' -c:v 'libx264' -bufsize, '64M' -maxrate '4M' -rtsp_transport 'tcp' -f 'rtsp' rtsp://localhost:31415/live.stream
*/

func ReadFromFileLoop(filename string, out io.Writer) {
	for {
		ReadFromFile(filename, out)
	}
}

func ReadFromFile(filename string, out io.Writer) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for {
		packet, err := liveview.ReadPacket(f)
		if err != nil {
			return
		}

		if packet.CommonHeader.PayloadType == liveview.COMMON_HEADER_TYPE_IMAGE {
			binary.Write(out, binary.BigEndian, packet.JpgData)
			log.Printf("Writing frame %d", packet.CommonHeader.SequenceNumber)
		}
	}
}

func StreamToRTSP(in io.Reader, rtspServerURL string) {
	// ffmpeg -re -f 'image2pipe' -i - -c:v 'libx264' -bufsize '64M' -maxrate '4M' -rtsp_transport 'tcp' -f 'rtsp' rtsp://localhost:8554/live.stream
	err := ffmpeg.
		Input("pipe:", ffmpeg.KwArgs{
			"re":          "",
			"stream_loop": "-1",
			"f":           "image2pipe",
		}).
		Output(rtspServerURL, ffmpeg.KwArgs{
			"c:v":            "libx264",
			"f":              "rtsp",
			"rtsp_transport": "tcp",
		}).
		WithInput(in).
		Run()

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	serverURL := "rtsp://localhost:8554/live.stream"
	pipeOut, pipeIn := io.Pipe()

	filename := "media/example-stream.bin"
	go ReadFromFileLoop(filename, pipeIn)
	StreamToRTSP(pipeOut, serverURL)
}
