package main

import (
	"fmt"
)

// Testing: sudo rtl_sdr -f 914000000 -s 2000000 -g -1 - | ./dump1090 --ifile -

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
	status = bladerf_set_txvga1(dev, -4);
printf("bladerf_set_txvga1\n");

	status = bladerf_set_txvga2(dev, 8);
printf("bladerf_set_txvga2\n");


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

func main() {
	bladeRFInit()
	v := make([]iq, 10000)
	for i := 0; i < 10000; i++ {
		if i%2 == 0 {
			v[i].i = 2047
			v[i].q = 0
		} else {
			v[i].i = 0
			v[i].q = 0
		}
	}
	for {
		bladeRFTX(v)
	}
}
