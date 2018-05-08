// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

import (
	"time"

	"periph.io/x/periph/host/cpu"
)

// ReadTime returns the time on a monotonic timer.
//
// It only works if allwinner-dma successfully loaded. Otherwise it returns 0.
func ReadTime() time.Duration {
	if drvDMA.timerMemory == nil {
		return 0
	}
	v := uint64(drvDMA.timerMemory.counterHigh)<<32 | uint64(drvDMA.timerMemory.counterLow)
	if v == 0 {
		// BUG(maruel): Implement using AVS_CNT0_REG on A64.
		return 0
	}
	// BUG(maruel): Assumes that counterCtrl & timerPLL6 is not set.
	const tick = time.Microsecond / 24
	return time.Duration(v) * tick
}

// Nanospin spins the CPU without calling into the kernel code if possible.
func Nanospin(t time.Duration) {
	start := ReadTime()
	if start == 0 {
		// Use the slow generic version.
		cpu.Nanospin(t)
		return
	}
	for ReadTime()-start < t {
	}
}

//

const (
	// 31:3 reserved
	timerPLL6            timerCtrl = 2 << 1 // CONT64_CLK_SRC_SEL; OSC24M if not set;
	timerReadLatchEnable timerCtrl = 1 << 1 // CONT64_RLATCH_EN; 1 to latch the counter to the registers
	timerClear                     = 1 << 0 // CONT64_CLR_EN; clears the counter
)

// R8: Page 96
type timerCtrl uint32

// timerMap is the mapping of important registers across CPUs.
type timerMap struct {
	reserved0   [0x80 / 4]uint32 //
	cntCtl      timerCtrl        // 0x80 AVS_CNT_CTL_REG AVS Control Register
	cnt0        uint32           // 0x84 AVS_CNT0_REG AVS Counter 0 Register
	cnt1        uint32           // 0x88 AVS_CNT1_REG AVS Counter 1 Register
	cndDrv      uint32           // 0x8C AVS_CNT_DIV_REG AVS Divisor Register
	reserved1   [0x10 / 4]uint32 // On R8 only.
	counterCtrl timerCtrl        // 0x0A0 COUNTER64_CTRL_REG 64-bit Counter control
	counterLow  uint32           // 0x0A4 COUNTER64_LOW_REG 64-bit Counter low
	counterHigh uint32           // 0x0A8 COUNTER64_HI_REG 64-bit Counter high
}

// A64: Page 161.
type timerMapA64 struct {
	reserved0  uint32    // 0x0 TMR_IRQ_EN_REG Timer IRQ Enable Register
	reserved1  uint32    // 0x4 TMR_IRQ_STA_REG Timer Status Register
	reserved2  uint32    // 0x10 TMR0_CTRL_REG Timer 0 Control Register
	reserved3  uint32    // 0x14 TMR0_INTV_VALUE_REG Timer 0 Interval Value Register
	reserved4  uint32    // 0x18 TMR0_CUR_VALUE_REG Timer 0 Current Value Register
	reserved5  uint32    // 0x20 TMR1_CTRL_REG Timer 1 Control Register
	reserved6  uint32    // 0x24 TMR1_INTV_VALUE_REG Timer 1 Interval Value Register
	reserved7  uint32    // 0x28 TMR1_CUR_VALUE_REG Timer 1 Current Value Register
	cntCtl     timerCtrl // 0x80 AVS_CNT_CTL_REG AVS Control Register
	cnt0       uint32    // 0x84 AVS_CNT0_REG AVS Counter 0 Register
	cnt1       uint32    // 0x88 AVS_CNT1_REG AVS Counter 1 Register
	cndDrv     uint32    // 0x8C AVS_CNT_DIV_REG AVS Divisor Register
	reserved8  uint32    // 0xA0 WDOG0_IRQ_EN_REG Watchdog 0 IRQ Enable Register
	reserved9  uint32    // 0xA4 WDOG0_IRQ_STA_REG Watchdog 0 Status Register
	reserved10 uint32    // 0xB0 WDOG0_CTRL_REG Watchdog 0 Control Register
	reserved11 uint32    // 0xB4 WDOG0_CFG_REG Watchdog 0 Configuration Register
	reserved12 uint32    // 0xB8 WDOG0_MODE_REG Watchdog 0 Mode Register
}

// R8: Page 85
type timerMapR8 struct {
	reserved0   uint32       // 0x000 ASYNC_TMR_IRQ_EN_REG Timer IRQ Enable
	reserved1   uint32       // 0x004 ASYNC_TMR_IRQ_STAS_REG Timer Status
	reserved2   [2]uint32    // 0x008-0x00C
	reserved3   uint32       // 0x010 ASYNC_TMR0_CTRL_REG Timer 0 Control
	reserved4   uint32       // 0x014 ASYNC_TMR0_INTV_VALUE_REG Timer 0 Interval Value
	reserved5   uint32       // 0x018 ASYNC_TMR0_CURNT_VALUE_REG Timer 0 Current Value
	reserved6   uint32       // 0x01C
	reserved7   uint32       // 0x020 ASYNC_TMR1_CTRL_REG Timer 1 Control
	reserved8   uint32       // 0x024 ASYNC_TMR1_INTV_VALUE_REG Timer 1 Interval Value
	reserved9   uint32       // 0x028 ASYNC_TMR1_CURNT_VALUE_REG Timer 1 Current Value
	reserved10  uint32       // 0x02C
	reserved11  uint32       // 0x030 ASYNC_TMR2_CTRL_REG Timer 2 Control
	reserved12  uint32       // 0x034 ASYNC_TMR2_INTV_VALUE_REG Timer 2 Interval Value
	reserved13  uint32       // 0x038 ASYNC_TMR2_CURNT_VALUE_REG Timer 2 Current Value
	reserved14  uint32       // 0x03C
	reserved15  uint32       // 0x040 ASYNC_TMR3_CTRL_REG Timer 3 Control
	reserved16  uint32       // 0x044 ASYNC_TMR3_INTV_VALUE_REG Timer 3 Interval Value
	reserved17  [2]uint32    // 0x048-0x04C
	reserved18  uint32       // 0x050 ASYNC_TMR4_CTRL_REG Timer 4 Control
	reserved19  uint32       // 0x054 ASYNC_TMR4_INTV_VALUE_REG Timer 4 Interval Value
	reserved20  uint32       // 0x058 ASYNC_TMR4_CURNT_VALUE_REG Timer 4 Current Value
	reserved21  uint32       // 0x05C
	reserved22  uint32       // 0x060 ASYNC_TMR5_CTRL_REG Timer 5 Control
	reserved23  uint32       // 0x064 ASYNC_TMR5_INTV_VALUE_REG Timer 5 Interval Value
	reserved24  uint32       // 0x068 ASYNC_TMR5_CURNT_VALUE_REG Timer 5 Current Value
	reserved25  [5]uint32    // 0x06C-0x07C
	cntCtl      timerCtrl    // 0x080 AVS_CNT_CTL_REG AVS Control Register
	cnt0        uint32       // 0x084 AVS_CNT0_REG AVS Counter 0 Register
	cnt1        uint32       // 0x088 AVS_CNT1_REG AVS Counter 1 Register
	cndDiv      uint32       // 0x08C AVS_CNT_DIVISOR_REG AVS Divisor
	reserved26  uint32       // 0x090 WDOG_CTRL_REG
	reserved27  uint32       // 0x094 WDOG_MODE_REG Watchdog Mode
	reserved28  [2]uint32    // 0x098-0x09C
	counterCtrl timerCtrl    // 0x0A0 COUNTER64_CTRL_REG 64-bit Counter control
	counterLow  uint32       // 0x0A4 COUNTER64_LOW_REG 64-bit Counter low
	counterHigh uint32       // 0x0A8 COUNTER64_HI_REG 64-bit Counter high
	reserved29  [0x94]uint32 // 0x0AC-0x13C
	reserved30  uint32       // 0x140 CPU_CFG_REG CPU configuration register
}
