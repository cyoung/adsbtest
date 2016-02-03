package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

/*
*8da826f558b5027c79975332ba18;
CRC: 000000 (ok)
DF 17: ADS-B message.
  Capability     : 5 (Level 2+3+4 (DF0,4,5,11,20,21,24,code7 - is airborne))
  ICAO Address   : a826f5
  Extended Squitter  Type: 11
  Extended Squitter  Sub : 0
  Extended Squitter  Name: Airborne Position (Baro Altitude)
    F flag   : even
    T flag   : non-UTC
    Altitude : 35000 feet
    Latitude : 81468 (not decoded)
    Longitude: 104275 (not decoded)
*/

// 000010204bc7e3

const (
	ADSB_LONG_LEN  = 14 // Bytes.
	ADSB_SHORT_LEN = 7  // Bytes.
)

func decodeDump1090Fmt(s string) ([]byte, error) {
	p := strings.Trim(s, "*;")

	// Check to make sure we have either 112 bits or 56 bits in the message.
	if len(p) != 2*ADSB_LONG_LEN && len(p) != 2*ADSB_SHORT_LEN {
		return nil, errors.New(fmt.Sprintf("invalid length frame (%d): %s", len(p), p))
	}

	// Decode to bytes.

	b, err := hex.DecodeString(p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func encodeBit(b uint8) []byte {
	ret := make([]byte, 2)
	if b != 0 {
		ret[0] = 1
	} else {
		ret[1] = 1
	}
	return ret
}

// Each byte returned represents 0.5µs in time domain.
func createPacket(packet []byte) []byte {
	// 8µs preamble. 16 bits.
	ret := make([]byte, 16)
	ret[0] = 1
	ret[1] = 0
	ret[2] = 1
	// ...
	ret[7] = 1
	ret[8] = 0
	ret[9] = 1

	// Now we translate each bit of 'packet' into a 1µs code.
	// A '10' in ret is a 1 in packet, a '01' in ret is a 0 in packet.

	for i := 0; i < len(packet); i++ {
		ret = append(ret, encodeBit((uint8(packet[i])&0x80)>>7)...)
		ret = append(ret, encodeBit((uint8(packet[i])&0x40)>>6)...)
		ret = append(ret, encodeBit((uint8(packet[i])&0x20)>>5)...)
		ret = append(ret, encodeBit((uint8(packet[i])&0x10)>>4)...)
		ret = append(ret, encodeBit((uint8(packet[i])&0x08)>>3)...)
		ret = append(ret, encodeBit((uint8(packet[i])&0x04)>>2)...)
		ret = append(ret, encodeBit((uint8(packet[i])&0x02)>>1)...)
		ret = append(ret, encodeBit(uint8(packet[i])&0x01)...)
	}
	return ret
}

func iqFileOut(packet []byte) []byte {
	ret := make([]byte, 0)
	for i := 0; i < len(packet); i++ {
		iq := make([]byte, 2)
		if packet[i] != 0 {
			iq[0] = 50
			iq[1] = 50
		} else {
			iq[0] = 127
			iq[1] = 127
		}
		ret = append(ret, iq...)
	}
	return ret
}

func main() {
	testMessage := "*8da826f558b5027c79975332ba18;"
	f, err := decodeDump1090Fmt(testMessage)
	fmt.Printf("%s\n", hex.Dump(f))
	if err != nil {
		panic(err)
	}
	p := createPacket(f)

	fmt.Printf("%d\n", len(p))

	fmt.Printf("%s\n", hex.Dump(p))
	fOut, err := os.Create("1090es.bin")
	if err != nil {
		panic(err)
	}
	defer fOut.Close()
	iq := iqFileOut(p)
	fmt.Printf("len=%d\n", len(iq))
	fmt.Printf("%s\n", hex.Dump(iq))
	//	for i := 0; i < 1000; i++ {
	fOut.Write(iq)
	//	}
}
