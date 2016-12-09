// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

const (
	spiClkEnable clockSPI = 1 << 31 // SCLK_GATING
	// 30:26 reserved
	spiClkOSC24M clockSPI = 0 << 24 // CLK_SRC_SEL
	spiClkPLL6   clockSPI = 1 << 24 // A64: PLL_PERIPH0(1X)
	spiClkPLL5   clockSPI = 2 << 24 // A64: PLL_PERIPH1(1X)  R8: PLL5 = DDR
	// 23:18 reserved
	spiClkDiv1a clockSPI = 0 << 16 // CLK_DIV_RATIO_N
	spiClkDiv2a clockSPI = 1 << 16 //
	spiClkDiv4a clockSPI = 2 << 16 //
	spiClkDiv8a clockSPI = 3 << 16 //
	// 15:4 reserved
	spiClkDiv1b  clockSPI = 0 << 0  // CLK_DIV_RATIO_M
	spiClkDiv2b  clockSPI = 1 << 0  //
	spiClkDiv3b  clockSPI = 2 << 0  //
	spiClkDiv4b  clockSPI = 3 << 0  //
	spiClkDiv5b  clockSPI = 4 << 0  //
	spiClkDiv6b  clockSPI = 5 << 0  //
	spiClkDiv7b  clockSPI = 6 << 0  //
	spiClkDiv8b  clockSPI = 7 << 0  //
	spiClkDiv9b  clockSPI = 8 << 0  //
	spiClkDiv10b clockSPI = 9 << 0  //
	spiClkDiv11b clockSPI = 10 << 0 //
	spiClkDiv12b clockSPI = 11 << 0 //
	spiClkDiv13b clockSPI = 12 << 0 //
	spiClkDiv14b clockSPI = 13 << 0 //
	spiClkDiv15b clockSPI = 14 << 0 //
	spiClkDiv16b clockSPI = 15 << 0 //
)

// Also valid for IR.
//
// SPI0_SCLK_CFG_REG / SPI1_SCLK_CFG_REG / SPI2_SCLK_CFG_REG / IR_SCLK_CFG_REG
//
// A64: Page 110-111. (Also Page 554?)
// R8: Page 71.
type clockSPI uint32

const (
	pll6Enable     pll6R8Ctl = 1 << 31 // PLL6_Enable
	pll6Force24Mhz pll6R8Ctl = 1 << 30 // PLL6_BYPASS_EN; force 24Mhz
	// 29:13 reserved
	pll6FactorMulN0  pll6R8Ctl = 0 << 8  // PLL6_FACTOR_N
	pll6FactorMulN1  pll6R8Ctl = 1 << 8  //
	pll6FactorMulN2  pll6R8Ctl = 2 << 8  //
	pll6FactorMulN3  pll6R8Ctl = 3 << 8  //
	pll6FactorMulN4  pll6R8Ctl = 4 << 8  //
	pll6FactorMulN5  pll6R8Ctl = 5 << 8  //
	pll6FactorMulN6  pll6R8Ctl = 6 << 8  //
	pll6FactorMulN7  pll6R8Ctl = 7 << 8  //
	pll6FactorMulN8  pll6R8Ctl = 8 << 8  //
	pll6FactorMulN9  pll6R8Ctl = 9 << 8  //
	pll6FactorMulN10 pll6R8Ctl = 10 << 8 //
	pll6FactorMulN11 pll6R8Ctl = 11 << 8 //
	pll6FactorMulN12 pll6R8Ctl = 12 << 8 //
	pll6FactorMulN13 pll6R8Ctl = 13 << 8 //
	pll6FactorMulN14 pll6R8Ctl = 14 << 8 //
	pll6FactorMulN15 pll6R8Ctl = 15 << 8 //
	pll6FactorMulN16 pll6R8Ctl = 16 << 8 //
	pll6FactorMulN17 pll6R8Ctl = 17 << 8 //
	pll6FactorMulN18 pll6R8Ctl = 18 << 8 //
	pll6FactorMulN19 pll6R8Ctl = 19 << 8 //
	pll6FactorMulN20 pll6R8Ctl = 20 << 8 //
	pll6FactorMulN21 pll6R8Ctl = 21 << 8 //
	pll6FactorMulN22 pll6R8Ctl = 22 << 8 //
	pll6FactorMulN23 pll6R8Ctl = 23 << 8 //
	pll6FactorMulN24 pll6R8Ctl = 24 << 8 //
	pll6FactorMulN25 pll6R8Ctl = 25 << 8 //
	pll6FactorMulN26 pll6R8Ctl = 26 << 8 //
	pll6FactorMulN27 pll6R8Ctl = 27 << 8 //
	pll6FactorMulN28 pll6R8Ctl = 28 << 8 //
	pll6FactorMulN29 pll6R8Ctl = 29 << 8 //
	pll6FactorMulN30 pll6R8Ctl = 30 << 8 //
	pll6FactorMulN31 pll6R8Ctl = 31 << 8 //
	pll6Damping      pll6R8Ctl = 2 << 6  //
	pll6FactorMulK1  pll6R8Ctl = 0 << 4  // PLL6_FACTOR_K
	pll6FactorMulK2  pll6R8Ctl = 1 << 4  //
	pll6FactorMulK3  pll6R8Ctl = 2 << 4  //
	pll6FactorMulK4  pll6R8Ctl = 3 << 4  //
	// 3:2 reserved
	pll6FactorDivM1 pll6R8Ctl = 0 << 4 // PLL6_FACTOR_M
	pll6FactorDivM2 pll6R8Ctl = 1 << 4 //
	pll6FactorDivM3 pll6R8Ctl = 2 << 4 //
	pll6FactorDivM4 pll6R8Ctl = 3 << 4 //
)

// PLL6_CFG_REG
// R8: Page 63; default 0x21009931
//
// Output = (24MHz*N*K)/M/2
// Note: the output 24MHz*N*K clock must be in the range of 240MHz~3GHz if the
// bypass is disabled.
type pll6R8Ctl uint32
