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

/*
#cgo linux LDFLAGS: -lbladeRF
#cgo darwin LDFLAGS: -lbladeRF
#include <stdio.h>
#include <libbladeRF.h>


struct bladerf *dev;

void bladeRFDeinit() {
	bladerf_close(dev);
}

int bladeRFInit() {
	int status = 0;
	// Open device.
	status = bladerf_open(&dev, "");
printf("bladerf_open\n");
	if (status != 0) return status;

	// TX bw - 1.5 MHz.
	int setBw = 0;
	status = bladerf_set_bandwidth(dev, BLADERF_MODULE_TX, 1500000, &setBw);
printf("bladerf_set_bandwidth\n");
	if (status != 0) return status;

	// Set frequency to 915 MHz.
	status = bladerf_set_frequency(dev, BLADERF_MODULE_TX, 915000000);
printf("bladerf_set_frequency\n");
	if (status != 0) return status;

	// Set 2 MHz samplerate.
	int setSampleRate = 0;
	status = bladerf_set_sample_rate(dev, BLADERF_MODULE_TX, 2000000, &setSampleRate);
printf("bladerf_set_sample_rate\n");
	if (status != 0) return status;

	// Configure TX module.
	const unsigned int num_buffers   = 16;
	const unsigned int buffer_size   = 8192;  // Must be a multiple of 1024.
	const unsigned int num_transfers = 8;
	const unsigned int timeout_ms    = 3500;

	status = bladerf_sync_config(dev,
								BLADERF_MODULE_TX,
								BLADERF_FORMAT_SC16_Q11,
								num_buffers,
								buffer_size,
								num_transfers,
								timeout_ms);
printf("bladerf_sync_config\n");

	if (status != 0) return status;

	// Enable the TX module.
	status = bladerf_enable_module(dev, BLADERF_MODULE_TX, true);
printf("bladerf_enable_module\n");

	if (status != 0) return status;

	// Set txvga1 (post-LPF gain) to -18dB.
	status = bladerf_set_txvga1(dev, -18);
printf("bladerf_set_txvga1\n");


	return status;

}

int bladeRFTX(int16_t *buf, int len) {
	return bladerf_sync_tx(dev, buf, len, NULL, 5000);
}
*/
import "C"
import "unsafe"

const (
	ADSB_LONG_LEN  = 14 // Bytes.
	ADSB_SHORT_LEN = 7  // Bytes.
)

// SC16Q11 for BladeRF.
type iq struct {
	i int16
	q int16
}

func bladeRFDeinit() {
	C.bladeRFDeinit()
}

func bladeRFInit() int {
	i := int(C.bladeRFInit())
	fmt.Printf("bladeRFInit=%d\n", i)
	return i
}

func bladeRFTX(samples []iq) int {
	buf := make([]int16, 2*len(samples))
	for k := 0; k < len(samples); k++ {
		buf[2*k] = samples[k].i
		buf[2*k+1] = samples[k].q
	}
	r := int(C.bladeRFTX((*C.int16_t)(unsafe.Pointer(&buf[0])), C.int(len(samples))))
	return r
}

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

// Fixed factor of 10x.
func interpolate(packet []byte) []byte {
	ret := make([]byte, 10*len(packet))
	for i := 0; i < len(packet); i++ {
		for j := 0; j < 10; j++ {
			ret[10*i+j] = packet[i]
		}
	}
	return ret
}

func iqPair(packet []byte) []iq {
	v := make([]iq, len(packet))
	for i := 0; i < len(packet); i++ {
		if packet[i] != 0 {
			v[i].i = 2040
			v[i].q = 0
		} else {
			v[i].i = 0
			v[i].q = 0
		}
	}
	return v
}

func iqOut(packet []byte) ([]byte, []iq) {
	// Add some white space to sync up on the other end for testing.
	spacing := make([]byte, 800)
	packet = append(packet, spacing...)
	//	packet = interpolate(packet)

	v := iqPair(packet)

	ret := make([]byte, len(packet)*4)
	for i := 0; i < len(v); i++ {
		ret[4*i] = byte(v[i].i >> 8)
		ret[4*i+1] = byte(v[i].i & 0xFF)
		ret[4*i+2] = byte(v[i].q >> 8)
		ret[4*i+3] = byte(v[i].q & 0xFF)
	}
	return ret, v
}

func main() {
	bladeRFInit()

	testMessage := "*8da826f558b5027c79975332ba18;"
	f, err := decodeDump1090Fmt(testMessage)
	fmt.Printf("%s\n", hex.Dump(f))
	if err != nil {
		panic(err)
	}
	p := createPacket(f)

	fmt.Printf("%d\n", len(p))

	//	fmt.Printf("%s\n", hex.Dump(p))
	fOut, err := os.Create("1090es.bin")
	if err != nil {
		panic(err)
	}
	defer fOut.Close()

	byteBuf, iq := iqOut(p)
	fmt.Printf("len=%d\n", len(byteBuf))
	//	fmt.Printf("%s\n", hex.Dump(byteBuf))

	for i := 0; i < 10000; i++ {
		if i%100 == 0 {
			fmt.Printf(".")
		}
		bladeRFTX(iq)
		fOut.Write(byteBuf)
	}
	fmt.Printf("\n")
	bladeRFDeinit()
}
