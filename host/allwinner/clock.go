// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

const (
	clockSPIEnable clockSPI = 1 << 31 // SCLK_GATING
	// 30:26 reserved
	clockSPIOSC24M clockSPI = 0 << 24 // CLK_SRC_SEL
	clockSPIPLL6   clockSPI = 1 << 24 // A64: PLL_PERIPH0(1X)
	clockSPIPLL5   clockSPI = 2 << 24 // A64: PLL_PERIPH1(1X)  R8: PLL5 = DDR
	// 23:18 reserved
	clockSPIDiv1a clockSPI = 0 << 16 // CLK_DIV_RATIO_N
	clockSPIDiv2a clockSPI = 1 << 16 //
	clockSPIDiv4a clockSPI = 2 << 16 //
	clockSPIDiv8a clockSPI = 3 << 16 //
	// 15:4 reserved
	clockSPIDiv1b  clockSPI = 0 << 0  // CLK_DIV_RATIO_M
	clockSPIDiv2b  clockSPI = 1 << 0  //
	clockSPIDiv3b  clockSPI = 2 << 0  //
	clockSPIDiv4b  clockSPI = 3 << 0  //
	clockSPIDiv5b  clockSPI = 4 << 0  //
	clockSPIDiv6b  clockSPI = 5 << 0  //
	clockSPIDiv7b  clockSPI = 6 << 0  //
	clockSPIDiv8b  clockSPI = 7 << 0  //
	clockSPIDiv9b  clockSPI = 8 << 0  //
	clockSPIDiv10b clockSPI = 9 << 0  //
	clockSPIDiv11b clockSPI = 10 << 0 //
	clockSPIDiv12b clockSPI = 11 << 0 //
	clockSPIDiv13b clockSPI = 12 << 0 //
	clockSPIDiv14b clockSPI = 13 << 0 //
	clockSPIDiv15b clockSPI = 14 << 0 //
	clockSPIDiv16b clockSPI = 15 << 0 //
)

// Also valid for IR.
//
// SPI0_SCLK_CFG_REG / SPI1_SCLK_CFG_REG / SPI2_SCLK_CFG_REG / IR_SCLK_CFG_REG
//
// A64: Page 110-111. (Also Page 554?)
// R8: Page 71.
type clockSPI uint32

const (
	clockPLL6Enable     clockPLL6R8Ctl = 1 << 31 // PLL6_Enable
	clockPLL6Force24Mhz clockPLL6R8Ctl = 1 << 30 // PLL6_BYPASS_EN; force 24Mhz
	// 29:13 reserved
	clockPLL6FactorMulN0  clockPLL6R8Ctl = 0 << 8  // PLL6_FACTOR_N
	clockPLL6FactorMulN1  clockPLL6R8Ctl = 1 << 8  //
	clockPLL6FactorMulN2  clockPLL6R8Ctl = 2 << 8  //
	clockPLL6FactorMulN3  clockPLL6R8Ctl = 3 << 8  //
	clockPLL6FactorMulN4  clockPLL6R8Ctl = 4 << 8  //
	clockPLL6FactorMulN5  clockPLL6R8Ctl = 5 << 8  //
	clockPLL6FactorMulN6  clockPLL6R8Ctl = 6 << 8  //
	clockPLL6FactorMulN7  clockPLL6R8Ctl = 7 << 8  //
	clockPLL6FactorMulN8  clockPLL6R8Ctl = 8 << 8  //
	clockPLL6FactorMulN9  clockPLL6R8Ctl = 9 << 8  //
	clockPLL6FactorMulN10 clockPLL6R8Ctl = 10 << 8 //
	clockPLL6FactorMulN11 clockPLL6R8Ctl = 11 << 8 //
	clockPLL6FactorMulN12 clockPLL6R8Ctl = 12 << 8 //
	clockPLL6FactorMulN13 clockPLL6R8Ctl = 13 << 8 //
	clockPLL6FactorMulN14 clockPLL6R8Ctl = 14 << 8 //
	clockPLL6FactorMulN15 clockPLL6R8Ctl = 15 << 8 //
	clockPLL6FactorMulN16 clockPLL6R8Ctl = 16 << 8 //
	clockPLL6FactorMulN17 clockPLL6R8Ctl = 17 << 8 //
	clockPLL6FactorMulN18 clockPLL6R8Ctl = 18 << 8 //
	clockPLL6FactorMulN19 clockPLL6R8Ctl = 19 << 8 //
	clockPLL6FactorMulN20 clockPLL6R8Ctl = 20 << 8 //
	clockPLL6FactorMulN21 clockPLL6R8Ctl = 21 << 8 //
	clockPLL6FactorMulN22 clockPLL6R8Ctl = 22 << 8 //
	clockPLL6FactorMulN23 clockPLL6R8Ctl = 23 << 8 //
	clockPLL6FactorMulN24 clockPLL6R8Ctl = 24 << 8 //
	clockPLL6FactorMulN25 clockPLL6R8Ctl = 25 << 8 //
	clockPLL6FactorMulN26 clockPLL6R8Ctl = 26 << 8 //
	clockPLL6FactorMulN27 clockPLL6R8Ctl = 27 << 8 //
	clockPLL6FactorMulN28 clockPLL6R8Ctl = 28 << 8 //
	clockPLL6FactorMulN29 clockPLL6R8Ctl = 29 << 8 //
	clockPLL6FactorMulN30 clockPLL6R8Ctl = 30 << 8 //
	clockPLL6FactorMulN31 clockPLL6R8Ctl = 31 << 8 //
	clockPLL6Damping      clockPLL6R8Ctl = 2 << 6  //
	clockPLL6FactorMulK1  clockPLL6R8Ctl = 0 << 4  // PLL6_FACTOR_K
	clockPLL6FactorMulK2  clockPLL6R8Ctl = 1 << 4  //
	clockPLL6FactorMulK3  clockPLL6R8Ctl = 2 << 4  //
	clockPLL6FactorMulK4  clockPLL6R8Ctl = 3 << 4  //
	// 3:2 reserved
	clockPLL6FactorDivM1 clockPLL6R8Ctl = 0 << 4 // PLL6_FACTOR_M
	clockPLL6FactorDivM2 clockPLL6R8Ctl = 1 << 4 //
	clockPLL6FactorDivM3 clockPLL6R8Ctl = 2 << 4 //
	clockPLL6FactorDivM4 clockPLL6R8Ctl = 3 << 4 //
)

// PLL6_CFG_REG
// R8: Page 63; default 0x21009931
//
// Output = (24MHz*N*K)/M/2
// Note: the output 24MHz*N*K clock must be in the range of 240MHz~3GHz if the
// bypass is disabled.
type clockPLL6R8Ctl uint32

// clockMap is the mapping of important registers across CPUs.
type clockMap struct {
	reserved0 [0xA0 / 4]uint32 //
	spi0Clk   clockSPI         // 0x0A0 SPI0_SCLK_CFG_REG SPI0 Clock
	spi1Clk   clockSPI         // 0x0A4 SPI1_SCLK_CFG_REG SPI1 Clock
	spi2Clk   clockSPI         // 0x0A8 SPI2_SCLK_CFG_REG SPI2 Clock (Not on A64)
}

// R8: Page 57-59.
type clockMapR8 struct {
	r0      uint32         // 0x000 PLL1_CFG_REG PLL1 Control
	r1      uint32         // 0x004 PLL1_TUN_REG PLL1 Tuning
	r2      uint32         // 0x008 PLL2_CFG_REG PLL2 Control
	r3      uint32         // 0x00C PLL2_TUN_REG PLL2 Tuning
	r4      uint32         // 0x010 PLL3_CFG_REG PLL3 Control
	r5      uint32         // 0x014
	r6      uint32         // 0x018 PLL4_CFG_REG PLL4 Control
	r7      uint32         // 0x01C
	r8      uint32         // 0x020 PLL5_CFG_REG PLL5 Control
	r9      uint32         // 0x024 PLL5_TUN_REG PLL5 Tuning
	r10     clockPLL6R8Ctl // 0x028 PLL6_CFG_REG PLL6 Control
	r11     uint32         // 0x02C PLL6 Tuning
	r12     uint32         // 0x030 PLL7_CFG_REG
	r13     uint32         // 0x034
	r14     uint32         // 0x038 PLL1_TUN2_REG PLL1 Tuning2
	r15     uint32         // 0x03C PLL5_TUN2_REG PLL5 Tuning2
	r16     uint32         // 0x04C
	r17     uint32         // 0x050 OSC24M_CFG_REG OSC24M control
	r18     uint32         // 0x054 CPU_AHB_APB0_CFG_REG CPU, AHB And APB0 Divide Ratio
	r19     uint32         // 0x058 APB1_CLK_DIV_REG APB1 Clock Divider
	r20     uint32         // 0x05C AXI_GATING_REG AXI Module Clock Gating
	r21     uint32         // 0x060 AHB_GATING_REG0 AHB Module Clock Gating 0
	r22     uint32         // 0x064 AHB_GATING_REG1 AHB Module Clock Gating 1
	r23     uint32         // 0x068 APB0_GATING_REG APB0 Module Clock Gating
	r24     uint32         // 0x06C APB1_GATING_REG APB1 Module Clock Gating
	r25     uint32         // 0x080 NAND_SCLK_CFG_REG Nand Flash Clock
	r26     uint32         // 0x084
	r27     uint32         // 0x088 SD0_SCLK_CFG_REG SD0 Clock
	r28     uint32         // 0x08C SD1_SCLK_CFG_REG SD1 Clock
	r29     uint32         // 0x090 SD2_SCLK_CFG_REG SD2 Clock
	r30     uint32         // 0x094
	r31     uint32         // 0x098
	r32     uint32         // 0x09C CE_SCLK_CFG_REG Crypto Engine Clock
	spi0Clk clockSPI       // 0x0A0 SPI0_SCLK_CFG_REG SPI0 Clock
	spi1Clk clockSPI       // 0x0A4 SPI1_SCLK_CFG_REG SPI1 Clock
	spi2Clk clockSPI       // 0x0A8 SPI2_SCLK_CFG_REG SPI2 Clock
	r33     uint32         // 0x0AC
	irClk   clockSPI       // 0x0B0 IR_SCLK_CFG_REG IR Clock
	r34     uint32         // 0x0B4
	r35     uint32         // 0x0B8
	r36     uint32         // 0x0BC
	r37     uint32         // 0x0C0
	r38     uint32         // 0x0C4
	r39     uint32         // 0x0C8
	r40     uint32         // 0x0CC
	r41     uint32         // 0x0D0
	r42     uint32         // 0x0D4
	r43     uint32         // 0x100 DRAM_SCLK_CFG_REG DRAM Clock
	r44     uint32         // 0x104 BE_CFG_REG Display Engine Backend Clock
	r45     uint32         // 0x108
	r46     uint32         // 0x10C FE_CFG_REG Display Engine Front End Clock
	r47     uint32         // 0x110
	r48     uint32         // 0x114
	r49     uint32         // 0x118
	r50     uint32         // 0x11C
	r51     uint32         // 0x120
	r52     uint32         // 0x124
	r53     uint32         // 0x128
	r54     uint32         // 0x12C LCD_CH1_CFG_REG LCD Channel1 Clock
	r55     uint32         // 0x130
	r56     uint32         // 0x134 CSI_CFG_REG CSI Clock
	r57     uint32         // 0x138
	r58     uint32         // 0x13C VE_CFG_REG Video Engine Clock
	r59     uint32         // 0x140 AUDIO_CODEC_SCLK_CFG_REG Audio Codec Gating Special Clock
	r60     uint32         // 0x144 AVS_SCLK_CFG_REG AVS Gating Special Clock
	r61     uint32         // 0x148
	r62     uint32         // 0x14C
	r63     uint32         // 0x150
	r64     uint32         // 0x154 MALI_CLOCK_CFG_REG Mali400 Gating Special Clock
	r65     uint32         // 0x158
	r66     uint32         // 0x15C MBUS_SCLK_CFG_REG MBUS Gating Clock
	r67     uint32         // 0x160 IEP_SCLK_CFG_REG IEP Gating Clock
}

// A64: Page 81-84.
type clockMapA64 struct {
	r0      uint32   // 0x000 PLL_CPUX_CTRL_REG PLL_CPUX Control Register
	r1      uint32   // 0x008 PLL_AUDIO_CTRL_REG PLL_AUDIO Control Register
	r2      uint32   // 0x010 PLL_VIDEO0_CTRL_REG PLL_VIDEO0 Control Register
	r3      uint32   // 0x018 PLL_VE_CTRL_REG PLL_VE Control Register
	r4      uint32   // 0x020 PLL_DDR0_CTRL_REG PLL_DDR0 Control Register
	r5      uint32   // 0x028 PLL_PERIPH0_CTRL_REG PLL_PERIPH0 Control Register
	r6      uint32   // 0x02C PLL_PERIPH1_CTRL_REG PLL_PERIPH1 Control Register
	r7      uint32   // 0x030 PLL_VIDEO1_CTRL_REG PLL_VIDEO1 Control Register
	r8      uint32   // 0x038 PLL_GPU_CTRL_REG PLL_GPU Control Register
	r9      uint32   // 0x040 PLL_MIPI_CTRL_REG PLL_MIPI Control Register
	r10     uint32   // 0x044 PLL_HSIC_CTRL_REG PLL_HSIC Control Register
	r11     uint32   // 0x048 PLL_DE_CTRL_REG PLL_DE Control Register
	r12     uint32   // 0x04C PLL_DDR1_CTRL_REG PLL_DDR1 Control Register
	r13     uint32   // 0x050 CPU_AXI_CFG_REG CPUX/AXI Configuration Register
	r14     uint32   // 0x054 AHB1_APB1_CFG_REG AHB1/APB1 Configuration Register
	r15     uint32   // 0x058 APB2_CFG_REG APB2 Configuration Register
	r16     uint32   // 0x05C AHB2_CFG_REG AHB2 Configuration Register
	r17     uint32   // 0x060 BUS_CLK_GATING_REG0 Bus Clock Gating Register 0
	r18     uint32   // 0x064 BUS_CLK_GATING_REG1 Bus Clock Gating Register 1
	r19     uint32   // 0x068 BUS_CLK_GATING_REG2 Bus Clock Gating Register 2
	r20     uint32   // 0x06C BUS_CLK_GATING_REG3 Bus Clock Gating Register 3
	r21     uint32   // 0x070 BUS_CLK_GATING_REG4 Bus Clock Gating Register 4
	r22     uint32   // 0x074 THS_CLK_REG THS Clock Register
	r23     uint32   // 0x080 NAND_CLK_REG NAND Clock Register
	r24     uint32   // 0x088 SDMMC0_CLK_REG SDMMC0 Clock Register
	r25     uint32   // 0x08C SDMMC1_CLK_REG SDMMC1 Clock Register
	r26     uint32   // 0x090 SDMMC2_CLK_REG SDMMC2 Clock Register
	r27     uint32   // 0x098 TS_CLK_REG TS Clock Register
	r28     uint32   // 0x09C CE_CLK_REG CE Clock Register
	spi0Clk clockSPI // 0x0A0 SPI0_CLK_REG SPI0 Clock Register
	spi1Clk clockSPI // 0x0A4 SPI1_CLK_REG SPI1 Clock Register
	r29     uint32   // 0x0B0 I2S/PCM-0_CLK_REG I2S/PCM-0 Clock Register
	r30     uint32   // 0x0B4 I2S/PCM-1_CLK_REG I2S/PCM-1 Clock Register
	r31     uint32   // 0x0B8 I2S/PCM-2_CLK_REG I2S/PCM-2 Clock Register
	r32     uint32   // 0x0C0 SPDIF_CLK_REG SPDIF Clock Register
	r33     uint32   // 0x0CC USBPHY_CFG_REG USBPHY Configuration Register
	r34     uint32   // 0x0F4 DRAM_CFG_REG DRAM Configuration Register
	r35     uint32   // 0x0F8 PLL_DDR_CFG_REG PLL_DDR Configuration Register
	r36     uint32   // 0x0FC MBUS_RST_REG MBUS Reset Register
	r37     uint32   // 0x100 DRAM_CLK_GATING_REG DRAM Clock Gating Register
	r38     uint32   // 0x104 DE_CLK_REG DE Clock Register
	r39     uint32   // 0x118 TCON0_CLK_REG TCON0 Clock Register
	r40     uint32   // 0x11C TCON1_CLK_REG TCON1 Clock Register
	r41     uint32   // 0x124 DEINTERLACE_CLK_REG DEINTERLACE Clock Register
	r42     uint32   // 0x130 CSI_MISC_CLK_REG CSI_MISC Clock Register
	r43     uint32   // 0x134 CSI_CLK_REG CSI Clock Register
	r44     uint32   // 0x13C VE_CLK_REG VE Clock Register
	r45     uint32   // 0x140 AC_DIG_CLK_REG AC Digital Clock Register
	r46     uint32   // 0x144 AVS_CLK_REG AVS Clock Register
	r47     uint32   // 0x150 HDMI_CLK_REG HDMI Clock Register
	r48     uint32   // 0x154 HDMI_SLOW_CLK_REG HDMI Slow Clock Register
	r49     uint32   // 0x15C MBUS_CLK_REG MBUS Clock Register
	r50     uint32   // 0x168 MIPI_DSI_CLK_REG MIPI_DSI Clock Register
	r51     uint32   // 0x1A0 GPU_CLK_REG GPU Clock Register
	r52     uint32   // 0x200 PLL_STABLE_TIME_REG0 PLL Stable Time Register0
	r53     uint32   // 0x204 PLL_STABLE_TIME_REG1 PLL Stable Time Register1
	r54     uint32   // 0x21C PLL_PERIPH1_BIAS_REG PLL_PERIPH1 Bias Register
	r55     uint32   // 0x220 PLL_CPUX_BIAS_REG PLL_CPUX Bias Register
	r56     uint32   // 0x224 PLL_AUDIO_BIAS_REG PLL_AUDIO Bias Register
	r57     uint32   // 0x228 PLL_VIDEO0_BIAS_REG PLL_VIDEO0 Bias Register
	r58     uint32   // 0x22C PLL_VE_BIAS_REG PLL_VE Bias Register
	r59     uint32   // 0x230 PLL_DDR0_BIAS_REG PLL_DDR0 Bias Register
	r60     uint32   // 0x234 PLL_PERIPH0_BIAS_REG PLL_PERIPH0 Bias Register
	r61     uint32   // 0x238 PLL_VIDEO1_BIAS_REG PLL_VIDEO1 Bias Register
	r62     uint32   // 0x23C PLL_GPU_BIAS_REG PLL_GPU Bias Register
	r63     uint32   // 0x240 PLL_MIPI_BIAS_REG PLL_MIPI Bias Register
	r64     uint32   // 0x244 PLL_HSIC_BIAS_REG PLL_HSIC Bias Register
	r65     uint32   // 0x248 PLL_DE_BIAS_REG PLL_DE Bias Register
	r66     uint32   // 0x24C PLL_DDR1_BIAS_REG PLL_DDR1 Bias Register
	r67     uint32   // 0x250 PLL_CPUX_TUN_REG PLL_CPUX Tuning Register
	r68     uint32   // 0x260 PLL_DDR0_TUN_REG PLL_DDR0 Tuning Register
	r69     uint32   // 0x270 PLL_MIPI_TUN_REG PLL_MIPI Tuning Register
	r70     uint32   // 0x27C PLL_PERIPH1_PAT_CTRL_REG PLL_PERIPH1 Pattern Control Register
	r71     uint32   // 0x280 PLL_CPUX_PAT_CTRL_REG PLL_CPUX Pattern Control Register
	r72     uint32   // 0x284 PLL_AUDIO_PAT_CTRL_REG PLL_AUDIO Pattern Control Register
	r73     uint32   // 0x288 PLL_VIDEO0_PAT_CTRL_REG PLL_VIDEO0 Pattern Control Register
	r74     uint32   // 0x28C PLL_VE_PAT_CTRL_REG PLL_VE Pattern Control Register
	r75     uint32   // 0x290 PLL_DDR0_PAT_CTRL_REG PLL_DDR0 Pattern Control Register
	r76     uint32   // 0x298 PLL_VIDEO1_PAT_CTRL_REG PLL_VIDEO1 Pattern Control Register
	r77     uint32   // 0x29C PLL_GPU_PAT_CTRL_REG PLL_GPU Pattern Control Register
	r78     uint32   // 0x2A0 PLL_MIPI_PAT_CTRL_REG PLL_MIPI Pattern Control Register
	r79     uint32   // 0x2A4 PLL_HSIC_PAT_CTRL_REG PLL_HSIC Pattern Control Register
	r80     uint32   // 0x2A8 PLL_DE_PAT_CTRL_REG PLL_DE Pattern Control Register
	r81     uint32   // 0x2AC PLL_DDR1_PAT_CTRL_REG0 PLL_DDR1 Pattern Control Register0
	r82     uint32   // 0x2B0 PLL_DDR1_PAT_CTRL_REG1 PLL_DDR1 Pattern Control Register1
	r83     uint32   // 0x2C0 BUS_SOFT_RST_REG0 Bus Software Reset Register 0
	r84     uint32   // 0x2C4 BUS_SOFT_RST_REG1 Bus Software Reset Register 1
	r85     uint32   // 0x2C8 BUS_SOFT_RST_REG2 Bus Software Reset Register 2
	r86     uint32   // 0x2D0 BUS_SOFT_RST_REG3 Bus Software Reset Register 3
	r87     uint32   // 0x2D8 BUS_SOFT_RST_REG4 Bus Software Reset Register 4
	r88     uint32   // 0x2F0 CCM_SEC_SWITCH_REG CCM Security Switch Register
	r89     uint32   // 0x300 PS_CTRL_REG PS Control Register
	r90     uint32   // 0x304 PS_CNT_REG PS Counter Register
	r91     uint32   // 0x320 PLL_LOCK_CTRL_REG PLL Lock Control Register
}
