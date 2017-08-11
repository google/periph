// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pmem implements handling of physical memory for user space programs.
//
// To make things confusing, a modern computer has many view of the memory
// (address spaces):
//
// User
//
// User mode address space is the virtual address space that an application
// runs in.  It is generally a tad less than half the addressable memory, so on
// a 32 bits system, the addressable range is 1.9Gb. For 64 bits OS, it depends
// but it usually at least 3.5Gb. The memory is virtual and can be flushed to
// disk in the swap file unless individual pages are locked.
//
// Kernel
//
// Kernel address space is the virtual address space the kernel sees. It often
// can see the currently active user space program on the current CPU core in
// addition to all the memory the kernel sees. The kernel memory pages that are
// not mlock()'ed are 'virtual' and can be flushed to disk in the swap file
// when there's not enough RAM available. On linux systems, the kernel
// addressed memory can be mapped in user space via `/dev/kmem`.
//
// Physical
//
// Physical memory address space is the actual address of each page in the DRAM
// chip and anything connected to the memory controller. The mapping may be
// different depending on what controller looks at the bus, like with IOMMU. So
// a peripheral (GPU, DMA controller) may have a different view of the physical
// memory than the host CPU. On linux systems, this memory can be mapped in
// user space via `/dev/mem`.
//
// CPU
//
// The CPU or its subsystems may memory map registers (for example, to control
// GPIO pins, clock speed, etc). This is not "real" memory, this is a view of
// registers but it still follows "mostly" the same semantic as DRAM backed
// physical memory.
//
// Some CPU memory may have very special semantic where the mere fact of
// reading has side effects. For example reading a specific register may
// latches another.
//
// CPU memory accesses are layered with multiple caches, usually named L1, L2
// and optionally L3. Some controllers (DMA) can see some cache levels (L2) but
// not others (L1) on some CPU architecture (bcm283x). This means that a user
// space program writing data to a memory page and immediately asking the DMA
// controller to read it may cause stale data to be read!
//
// Hypervisor
//
// Hypervisor can change the complete memory mapping as seen by the kernel.
// This is outside the scope of this project. :)
//
// Summary
//
// In practice, the semantics change between CPU manufacturers (Broadcom vs
// Allwinner) and between architectures (ARM vs x86). The most tricky one is to
// understand cached memory and how it affects coherence and performance.
// Uncached memory is extremely slow so it must only be used when necessary.
//
// References
//
// Overview of IOMMU:
// https://en.wikipedia.org/wiki/Input-output_memory_management_unit
package pmem
