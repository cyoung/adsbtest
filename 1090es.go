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

// A '10' in ret is a 1 in packet, a '01' in ret is a 0 in packet.
func encodeBit(z uint8) int16 {
	return int16(1 << z)
}

// Each bit represents 0.5µs in time domain.
func createPacket(packet []byte) []int16 {
	// 8µs preamble. 16 bits.
	ret := make([]int16, 1)
	ret[0] = -0x5f60 // 0xA0A0

	// Now we translate each bit of 'packet' into a 1µs code.
	for i := 0; i < len(packet); i++ {
		enc := int16(int32(encodeBit((uint8(packet[i])&0x80)>>7)) << 14)
		enc = enc | (encodeBit((uint8(packet[i])&0x40)>>6) << 12)
		enc = enc | (encodeBit((uint8(packet[i])&0x20)>>5) << 10)
		enc = enc | (encodeBit((uint8(packet[i])&0x10)>>4) << 8)
		enc = enc | (encodeBit((uint8(packet[i])&0x08)>>3) << 6)
		enc = enc | (encodeBit((uint8(packet[i])&0x04)>>2) << 4)
		enc = enc | (encodeBit((uint8(packet[i])&0x02)>>1) << 2)
		enc = enc | (encodeBit(uint8(packet[i]) & 0x01))
		ret = append(ret, enc)
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

	j := make([]byte, 2*len(p))
	for i := 0; i < len(p); i++ {
		j[2*i] = byte(p[i] >> 8)
		j[(2*i)+1] = byte(p[i] & 0xFF)
	}

	fmt.Printf("%d\n", len(p))

	fmt.Printf("%s\n", hex.Dump(j))
	fOut, err := os.Create("1090es.bin")
	if err != nil {
		panic(err)
	}
	defer fOut.Close()
	fOut.Write(j)
}
