// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package allwinner exposes the GPIO functionality that is common to all
// AllWinner processors.
//
// This driver implements memory-mapped GPIO pin manipulation and leverages
// sysfs-gpio for edge detection.
//
// If you are looking at the actual implementation, open doc.go for further
// implementation details.
//
// Datasheets
//
// A64: http://files.pine64.org/doc/datasheet/pine64/Allwinner_A64_User_Manual_V1.0.pdf
//
// H3: http://dl.linux-sunxi.org/H3/Allwinner_H3_Datasheet_V1.0.pdf
//
// R8: https://github.com/NextThingCo/CHIP-Hardware/raw/master/CHIP%5Bv1_0%5D/CHIPv1_0-BOM-Datasheets/Allwinner%20R8%20User%20Manual%20V1.1.pdf
//
// Physical overview: http://files.pine64.org/doc/datasheet/pine64/A64_Datasheet_V1.1.pdf
package allwinner

// Other implementation details
//
// The most active kernel branch is
// https://github.com/linux-sunxi/linux-sunxi/commits/sunxi-next
//
// In particular look at
// https://github.com/linux-sunxi/linux-sunxi/blob/sunxi-next/drivers/dma/sun4i-dma.c
// and
// https://github.com/linux-sunxi/linux-sunxi/blob/sunxi-next/drivers/dma/sun6i-dma.c
