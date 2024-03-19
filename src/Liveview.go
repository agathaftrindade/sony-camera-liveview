package liveview

import (
	"log"
	"net/http"
	"time"
)

func (s *SonyCamera) StartStreaming(targetFPS int, out chan *Packet) {
	go s.streamingJob(targetFPS, out)
}

func (s *SonyCamera) streamingJob(targetFPS int, out chan *Packet) {
	//TODO add reconnect and backoff

	res, err := http.Get(*s.LiveviewURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()

	then := time.Now()
	for {

		targetMS := int64(1000.0 * 1000.0 / targetFPS)
		elapsed := time.Since(then).Microseconds()
		if elapsed < targetMS {
			// log.Printf("[s] Sleeping for %d ms \n", targetMS-elapsed)
			// time.Sleep(time.Duration(targetMS-elapsed) * time.Microsecond)
		}

		packet, err := ReadPacket(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		out <- packet
		then = time.Now()
	}
}
