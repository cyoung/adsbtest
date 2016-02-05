package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
)

/*
31db57800c92ae60148006745f105011a02c31c9832db2cf4e5a832df0c2fcb7cb4833d70c342d4810d9336008b3b0cf5f5e741e00002d0eaac08210000000ff0c51b92000000000efd304011a1518011b0300c5aba371de58c598c33d2658c372631b8e58c430434ab658c5aba371de581e00002d0eaac08210000000ff0c51b72000000000efd304011a1518011b0300c5aba371de58c598c33d2658c372631b8e58c430434ab658c5aba371de582180067403503455014a02c15cd832df0c35cda8015543e0c35c30d4b520c704cd803312830cefc30801cf0cb481234b8013f2813310cb4ca079c114c30cb8c30c30f5e7402180067403503455014a02ca092832df0c35cda8015543e0c36c30d0b520c704cd803312830c6f370c60073c32da048d2e004fca04cc432d3781e704530c30db1c31c7d79d2180067403503455014a02c83d4832df0c35cda8015543e0cf5c30ccb520c704cd803312830def370ca0073c32d2048d2e004fca04cc432d3181e704530c37cb1c31dfd79d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
3c62ab89c854bb7000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
3c62ab89c854a470158000213f5a0082102c3305000831021e012c3305000000000000000fd900071d0d00071d1500c5cae3ae4800248000353f42782210000000ff003ec72c7c4d5060cb9c31c75833df2c33c757f2d77cb07d77cf7c14b7df731fc75c37cb9c31c757f1d70df2e70cb1d5fcb1c1fcb3c1f05f65f7df7de0002a8000353f22002210000000ff004aee8c7c4d5060cb8c77c30833db9e35db97f3d38d5f5df49fdf1c330dfc75c37cb8c77c307f1d70df3c70cf0c1fc32c1fc72c1f05f65f7f7c70cc37d2df1c330df4c13093814de0002a8000353f22002210000000ff004aee847c4d5060cb8c77c30833db9e35db77f3d38c1f5df49fdf1c3309fc75c37cb8c77c307f1d70df3c70cf0c1fc33c1fc72c1f05f65f7f7c70cc27d2df1c3309f4c13093814de0002a8000353f22002210000000ff004aee887c4d5060cb8c77c30833db9e35db87f3d37d5f5df49fdf1c3305fc75c37cb8c77c307f1d70df3c70cf0c1fc30c1fc72c1f05f65f7f7c70cc17d2df1c3305f4c13093814de00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
*/

/*

	Data rate: 1.041667Mbps

*/

/*
#cgo linux LDFLAGS: -L. -lschifra_reed_solomon -lstdc++ -lm
#cgo darwin LDFLAGS: -L. -lschifra_reed_solomon -lstdc++ -lm
#include <stdlib.h>
#include <stdint.h>
extern void doRS(char *buf);
*/
import "C"
import "unsafe"

func doRS(buf []byte) []byte {
	if len(buf) != 72 {
		panic("doRS(): input not right length.")
	}
	ret := make([]byte, 0)
	ret = append(ret, make([]byte, 163)...)

	ret = append(ret, buf...)
	C.doRS((*C.char)(unsafe.Pointer(&ret[0])))
	fmt.Printf("%d %v\n", len(ret), ret)
	ret = ret[235:]
	return ret
}

const (
	UAT_LONG_LEN = 432 // Bytes.
)

var UPLINK_SYNC_WORD = []bool{false, false, false, true, false, true, false, true, false, false, true, true, false, false, true, false, false, false, true, false, false, true, false, true, true, false, true, true, false, false, false, true, true, true, false, true}
var DOWNLINK_SYNC_WORD = []bool{true, true, true, false, true, false, true, false, true, true, false, false, true, true, false, true, true, true, false, true, true, false, true, false, false, true, false, false, true, true, true, false, false, false, true, false}

func decodeDumpFmt(s string) ([]byte, error) {
	// Check to make sure we have either 112 bits or 56 bits in the message.
	if len(s) != 2*UAT_LONG_LEN {
		return nil, errors.New(fmt.Sprintf("invalid length frame (%d): %s", len(s), s))
	}

	// Decode to bytes.

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func encodeBit(b uint8) bool {
	if b != 0 {
		return true
	}
	return false
}

type iq struct {
	i int8
	q int8
}

func encodePacket(packet []bool) []iq {
	ret := make([]iq, 2)

	for i := 0; i < len(packet); i++ {
		v := make([]iq, 2)

		var pos_iq iq
		var neg_iq iq
		pos_iq.i = int8(50.0*math.Cos(float64(i)*math.Pi/5) + 128)
		pos_iq.q = int8(50.0*math.Sin(float64(i)*math.Pi/5) + 128)
		neg_iq.i = int8(50.0*math.Sin(float64(i)*math.Pi/5) + 128)
		neg_iq.q = int8(-50.0*math.Cos(float64(i)*math.Pi/5) + 128)
		if packet[i] {
			v[0] = neg_iq
			v[1] = pos_iq
		} else {
			v[0] = pos_iq
			v[1] = neg_iq
		}

		ret = append(ret, v...)
	}
	return ret
}

// "12.4.4.2.2.3  Interleaving Procedure".
// Returns an array of bytes in transmission order.
func interleavePacket(slicedPacket [][]byte) []byte {
	// Assume that all of the internal slices are of the same size.
	//FIXME: error checking.
	ret := make([]byte, 0)
	for i := 0; i < 92; i++ {
		for j := 0; j < 6; j++ {
			ret = append(ret, slicedPacket[j][i])
		}
	}
	return ret
}

// Expand into bits (bool).
func createPacket(packet []byte) []iq {
	// 36 bit preamble.
	ret := UPLINK_SYNC_WORD

	// Split into 72-byte slices.
	slicedPacket := make([][]byte, 0)
	for i := 0; i < UAT_LONG_LEN; i += 72 {
		slicedPacket = append(slicedPacket, packet[i:i+72])
	}

	// Reed-Solomon, FEC. "14.4.4.2.2.2 FEC Parity (before interleaving and after de-interleaving)".
	for i := 0; i < len(slicedPacket); i++ {
		// Append the Reed-Solomon parity bits to the end of the 72 byte row.
		rs_p := doRS(slicedPacket[i])
		slicedPacket[i] = append(slicedPacket[i], rs_p...)
		fmt.Printf("%d (%d): %s\n", i, len(slicedPacket[i]), hex.Dump(slicedPacket[i]))
	}
	// Interleave the message.

	//	packet = interleavePacket(packet)
	packet = interleavePacket(slicedPacket)
	fmt.Printf("done-\n%s\n", hex.Dump(packet))
	// Now we translate each bit of 'packet' into a bool value.

	for i := 0; i < len(packet); i++ {
		ret = append(ret, encodeBit((uint8(packet[i])&0x80)>>7))
		ret = append(ret, encodeBit((uint8(packet[i])&0x40)>>6))
		ret = append(ret, encodeBit((uint8(packet[i])&0x20)>>5))
		ret = append(ret, encodeBit((uint8(packet[i])&0x10)>>4))
		ret = append(ret, encodeBit((uint8(packet[i])&0x08)>>3))
		ret = append(ret, encodeBit((uint8(packet[i])&0x04)>>2))
		ret = append(ret, encodeBit((uint8(packet[i])&0x02)>>1))
		ret = append(ret, encodeBit(uint8(packet[i])&0x01))
	}
	return encodePacket(ret)
}

// Create a []byte that can be output to a file.
func iqFileOut(packet []iq) []byte {
	ret := make([]byte, 0)
	for _, pkt := range packet {
		ret = append(ret, byte(pkt.i))
		ret = append(ret, byte(pkt.q))
	}
	return ret
}

func main() {
	ee, _ := hex.DecodeString("0dd90007150b3908050b39c51243b0b80005800021bcc09082102d53cc00082efc1e012d43cc000000000100000fd9000f1300120813000fc46743b25400158000213d1ed082102cb9526ade833b9ee34eb34a04fdb4a1c6d96586a8")
	fmt.Printf("a\n%s\n\n", hex.Dump(ee))
	eee := doRS(ee[:72])
	fmt.Printf("b\n%s\n\n", hex.Dump(eee))
	return
	testMessage := "31db57800c92ae60148006745f105011a02c31c9832db2cf4e5a832df0c2fcb7cb4833d70c342d4810d9336008b3b0cf5f5e741e00002d0eaac08210000000ff0c51b92000000000efd304011a1518011b0300c5aba371de58c598c33d2658c372631b8e58c430434ab658c5aba371de581e00002d0eaac08210000000ff0c51b72000000000efd304011a1518011b0300c5aba371de58c598c33d2658c372631b8e58c430434ab658c5aba371de582180067403503455014a02c15cd832df0c35cda8015543e0c35c30d4b520c704cd803312830cefc30801cf0cb481234b8013f2813310cb4ca079c114c30cb8c30c30f5e7402180067403503455014a02ca092832df0c35cda8015543e0c36c30d0b520c704cd803312830c6f370c60073c32da048d2e004fca04cc432d3781e704530c30db1c31c7d79d2180067403503455014a02c83d4832df0c35cda8015543e0cf5c30ccb520c704cd803312830def370ca0073c32d2048d2e004fca04cc432d3181e704530c37cb1c31dfd79d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	f, err := decodeDumpFmt(testMessage)
	fmt.Printf("%s\n", hex.Dump(f))
	if err != nil {
		panic(err)
	}
	p := createPacket(f)

	fmt.Printf("%d\n", len(p))

	fOut, err := os.Create("uat.bin")
	if err != nil {
		panic(err)
	}
	defer fOut.Close()
	outByte := iqFileOut(p)
	fmt.Printf("len=%d\n", len(outByte))
	fmt.Printf("%s\n", hex.Dump(outByte))
	for i := 0; i < 10000; i++ {
		fOut.Write(outByte)
	}
}
