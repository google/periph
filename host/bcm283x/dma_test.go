// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"reflect"
	"testing"
)

func TestDmaStatus_String(t *testing.T) {
	if s := dmaStatus(0).String(); s != "0" {
		t.Fatal(s)
	}
	d := ^dmaStatus(0)
	if s := d.String(); s != "Reset|Abort|DisableDebug|WaitForOutstandingWrites|ErrorStatus|WaitingForOutstandingWrites|DreqStopsDMA|Paused|Dreq|Interrupt|End|Active|pp15|p15|dmaStatus(0xf00fe80)" {
		t.Fatal(s)
	}
}

func TestDmaTransferInfo_String(t *testing.T) {
	if s := dmaTransferInfo(0).String(); s != "Fire" {
		t.Fatal(s)
	}
	d := ^dmaTransferInfo(0)
	if s := d.String(); s != "NoWideBursts|SrcIgnore|SrcDReq|SrcWidth128|SrcInc|DstIgnore|DstDReq|DstWidth128|DstInc|WaitResp|Transfer2DMode|InterruptEnable|waits=31|burst=15|SlimBusDC9|dmaTransferInfo(0xf8000004)" {
		t.Fatal(s)
	}
}

func TestDmaDebug_String(t *testing.T) {
	if s := dmaDebug(0).String(); s != "0" {
		t.Fatal(s)
	}
	d := ^dmaDebug(0)
	if s := d.String(); s != "Lite|ReadError|FIFOError|ReadLastNotSetError|v7|state(1ff)|#ff|OutstandingWrites=15|dmaDebug(0xe0000008)" {
		t.Fatal(s)
	}
}

func TestDmaStride_String(t *testing.T) {
	if s := dmaStride(0).String(); s != "0x0" {
		t.Fatal(s)
	}
	d := ^dmaStride(0)
	if s := d.String(); s != "0xffff,0xffff" {
		t.Fatal(s)
	}
}

func TestControlBlock(t *testing.T) {
	c := controlBlock{}
	if c.initBlock(0, 0, 0, true, true, false, false, dmaFire, 0) == nil {
		t.Fatal("can't set both")
	}
	if c.initBlock(0, 0, 0, false, false, true, true, dmaFire, 0) == nil {
		t.Fatal("need at least one addr")
	}
	if c.initBlock(0, 1, 0, true, false, false, true, dmaFire, 0) == nil {
		t.Fatal("srcIO requires srcAddr")
	}
	if c.initBlock(1, 0, 0, false, true, true, false, dmaFire, 0) == nil {
		t.Fatal("dstIO requires dstAddr")
	}
	if c.initBlock(1, 1, 0, false, false, true, true, dmaSrcIgnore, 0) == nil {
		t.Fatal("must not specify anything other than clock source")
	}
	if c.initBlock(1, 1, 0, false, false, true, true, dmaFire, 100) == nil {
		t.Fatal("dstIO requires dstAddr")
	}
	if c.initBlock(1, 1, 0, false, false, true, true, dmaFire, 1) == nil {
		t.Fatal("dmaFire can't use waits")
	}

	if err := c.initBlock(1, 0, 0, false, false, true, true, dmaFire, 0); err != nil {
		t.Fatal(err)
	}
	if err := c.initBlock(0, 1, 0, false, false, true, true, dmaFire, 0); err != nil {
		t.Fatal(err)
	}
	if err := c.initBlock(1, 0, 0, true, false, false, true, dmaFire, 0); err != nil {
		t.Fatal(err)
	}
	if err := c.initBlock(0, 1, 0, false, true, true, false, dmaPCMTX, 0); err != nil {
		t.Fatal(err)
	}
}

func TestControlBlockGo_String(t *testing.T) {
	c := controlBlock{}
	if err := c.initBlock(0, 1, 0, false, true, false, false, dmaPCMTX, 0); err != nil {
		t.Fatal(err)
	}
	expected := "{\n  transferInfo: NoWideBursts|SrcIgnore|DstDReq|WaitResp|PCMTX,\n  srcAddr:      0x0,\n  dstAddr:      0x7e000001,\n  txLen:        0,\n  stride:       0x0,\n  nextCB:       0x0,\n}"
	if s := c.GoString(); s != expected {
		t.Fatalf("%q", s)
	}
}

func TestDmaChannel(t *testing.T) {
	d := dmaChannel{}
	if !d.isAvailable() {
		t.Fatal("empty channel is available")
	}
	d = dmaChannel{cs: dmaEnd}
	if err := d.wait(); err != nil {
		t.Fatal(err)
	}
	d = dmaChannel{debug: dmaReadError}
	if d.wait() == nil {
		t.Fatal("read error")
	}
	d = dmaChannel{debug: dmaFIFOError}
	if d.wait() == nil {
		t.Fatal("fifo error")
	}
	d = dmaChannel{debug: dmaReadLastNotSetError}
	if d.wait() == nil {
		t.Fatal("read last not set error")
	}
}

func TestDmaChannel_GoString(t *testing.T) {
	d := dmaChannel{}
	d.reset()
	d.startIO(0)
	expected := "{\n  cs:           WaitForOutstandingWrites|Active|pp8|p8,\n  cbAddr:       0x0,\n  transferInfo: Fire,\n  srcAddr:      0x0,\n  dstAddr:      0x0,\n  txLen:        0,\n  stride:       0x0,\n  nextCB:       0x0,\n  debug:        0,\n  reserved:     {...},\n}"
	if s := d.GoString(); s != expected {
		t.Fatalf("%q", s)
	}
}

func TestDmaMap_GoString(t *testing.T) {
	d := dmaMap{}
	// I have to admit, this is the worst test ever.
	if s := d.GoString(); len(s) != 3629 {
		t.Fatal(s, len(s))
	}
}

func TestStructSizes(t *testing.T) {
	// Verify internal assumptions.
	if s := reflect.TypeOf((*controlBlock)(nil)).Elem().Size(); s != 256/8 {
		t.Fatalf("controlBlock size: %d", s)
	}
	if s := reflect.TypeOf((*dmaChannel)(nil)).Elem().Size(); s != 0x100 {
		t.Fatalf("dmaChannel size: %d", s)
	}

}
