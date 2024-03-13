package liveview

import (
	"encoding/binary"
	"errors"
	"io"
	"slices"
)

const COMMON_HEADER_START_BYTE = 255
const COMMON_HEADER_TYPE_IMAGE = 1
const COMMON_HEADER_TYPE_INFO = 2

/*
Start byte : 1 [B]

	0xFF, fixed

Payload type : 1 [B]

	0x01 = For liveview images
	0x02 = For Liveview Frame Information
	- indicates type of the Payload

Sequence number : 2 [B]

	Frame No, 2 bytes integer and increments every frame
	This frame no will be repeated.

Time stamp : 4 [B]

	4 bytes integer, the unit will be indicated by Payload type
	- In case Payload type = 0x01, the unit of the Time stamp of the Common Header is
	milliseconds. The start time may not start from zero and depends on the server.
*/
type CommonHeader struct {
	StartByte      uint8
	PayloadType    uint8
	SequenceNumber uint16
	Timestamp      uint32
}

type PayloadHeaderPrefix struct {
	StartCode    [4]byte
	JpegDataSize [3]byte
	PaddingSize  uint8
}

type PayloadHeaderImage struct {
	StartCode    [4]byte
	JpegDataSize [3]byte
	PaddingSize  uint8
	_            [4]byte
	Flag         uint8
	_            [115]byte
}

type PayloadHeaderInfo struct {
	StartCode           [4]byte
	FrameInfoData       [3]byte
	PaddingSize         uint8
	Version             uint16
	FrameCount          uint16
	SingleFrameDataSize uint16
	_                   [114]byte
}

type Packet struct {
	CommonHeader       *CommonHeader
	PayloadHeaderImage *PayloadHeaderImage
	PayloadHeaderInfo  *PayloadHeaderInfo
	JpgData            []byte
}

func threeByteArrayToUint32(b [3]byte) uint32 {
	padding := make([]byte, 1)
	s := slices.Concat(padding, b[:])
	return binary.BigEndian.Uint32(s)
}

func parseSection[K CommonHeader | PayloadHeaderImage | PayloadHeaderInfo | []byte](reader io.Reader, obj *K) error {
	err := binary.Read(reader, binary.BigEndian, obj)
	return err
}

func parseImageData(reader io.Reader, packet *Packet) error {
	ph := PayloadHeaderImage{}
	err := parseSection(reader, &ph)
	if err != nil {
		return err
	}

	PAYLOAD_HEADER_START_CODE := [4]byte{0x24, 0x35, 0x68, 0x79}
	if ph.StartCode != PAYLOAD_HEADER_START_CODE {
		err := errors.New("invalid image payload header")
		return err
	}

	jpgSize := threeByteArrayToUint32(ph.JpegDataSize)
	jpgData := make([]byte, jpgSize)
	err = parseSection(reader, &jpgData)
	if err != nil {
		return err
	}

	if ph.PaddingSize != 0 {
		p := make([]byte, ph.PaddingSize)
		err = parseSection(reader, &p)
		if err != nil {
			return err
		}
	}

	packet.PayloadHeaderImage = &ph
	packet.JpgData = jpgData

	return nil
}

func parseInfoData(reader io.Reader, packet *Packet) error {
	ph := PayloadHeaderInfo{}
	err := parseSection(reader, &ph)
	if err != nil {
		return err
	}

	PAYLOAD_HEADER_START_CODE := [4]byte{0x24, 0x35, 0x68, 0x79}
	if ph.StartCode != PAYLOAD_HEADER_START_CODE {
		err := errors.New("invalid image payload header")
		return err
	}

	//TODO parse?

	toSkip := ph.SingleFrameDataSize * ph.SingleFrameDataSize
	p := make([]byte, toSkip)
	err = parseSection(reader, &p)
	if err != nil {
		return err
	}

	packet.PayloadHeaderInfo = &ph

	return nil
}

func ReadPacket(reader io.Reader) (*Packet, error) {

	var packet Packet

	ch := CommonHeader{}
	err := parseSection(reader, &ch)
	if err != nil {
		return nil, err
	}

	if ch.StartByte != COMMON_HEADER_START_BYTE {
		err := errors.New("invalid common header")
		return nil, err
	}

	packet.CommonHeader = &ch

	switch ch.PayloadType {
	case COMMON_HEADER_TYPE_IMAGE:
		parseImageData(reader, &packet)

	case COMMON_HEADER_TYPE_INFO:
		parseInfoData(reader, &packet)
	}

	return &packet, nil
}
