/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-09 17:15:16
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-11 10:29:15
 * @FilePath: \go-toolbox\pkg\crc\config.go
 * @Description: CRC算法配置库
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package crc

// CRC算法的标准配置
var (
	// CRC-4/ITU标准配置
	CRC4_ITU = Config{
		Width:  4,
		Poly:   0x03,
		Init:   0x00,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x00,
	}

	// CRC-5/EPC标准配置
	CRC5_EPC = Config{
		Width:  5,
		Poly:   0x09,
		Init:   0x09,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-5/ITU标准配置
	CRC5_ITU = Config{
		Width:  5,
		Poly:   0x15,
		Init:   0x00,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x00,
	}

	// CRC-5/USB标准配置
	CRC5_USB = Config{
		Width:  5,
		Poly:   0x05,
		Init:   0x1F,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x1F,
	}

	// CRC-6/ITU标准配置
	CRC6_ITU = Config{
		Width:  6,
		Poly:   0x03,
		Init:   0x00,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x00,
	}

	// CRC-7/MMC标准配置
	CRC7_MMC = Config{
		Width:  7,
		Poly:   0x09,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-8标准配置
	CRC8 = Config{
		Width:  8,
		Poly:   0x07,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-8/ATM标准配置
	CRC8_ATM = Config{
		Width:  8,
		Poly:   0x07,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-8/ITU标准配置
	CRC8_ITU = Config{
		Width:  8,
		Poly:   0x07,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x55,
	}

	// CRC-8/ROHC标准配置
	CRC8_ROHC = Config{
		Width:  8,
		Poly:   0x07,
		Init:   0xFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x00,
	}

	// CRC-8/MAXIM标准配置
	CRC8_MAXIM = Config{
		Width:  8,
		Poly:   0x31,
		Init:   0x00,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x00,
	}

	// CRC-8/CDMA2000标准配置
	CRC8_CDMA2000 = Config{
		Width:  8,
		Poly:   0x9B,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-8/DALLAS/1-WIRE标准配置
	CRC8_DALLAS_1WIRE = Config{
		Width:  8,
		Poly:   0x31,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-16/IBM标准配置
	CRC16_IBM = Config{
		Width:  16,
		Poly:   0x8005,
		Init:   0x0000,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000,
	}

	// CRC-16/MAXIM标准配置
	CRC16_MAXIM = Config{
		Width:  16,
		Poly:   0x8005,
		Init:   0x0000,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFF,
	}

	// CRC-16/USB标准配置
	CRC16_USB = Config{
		Width:  16,
		Poly:   0x8005,
		Init:   0xFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFF,
	}

	// CRC-16/MODBUS标准配置
	CRC16_MODBUS = Config{
		Width:  16,
		Poly:   0x8005,
		Init:   0xFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000,
	}

	// CRC-16/CCITT标准配置
	CRC16_CCITT = Config{
		Width:  16,
		Poly:   0x1021,
		Init:   0xFFFF,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x0000,
	}

	// CRC-16/CCITT-FALSE标准配置
	CRC16_CCITT_FALSE = Config{
		Width:  16,
		Poly:   0x1021,
		Init:   0x0000,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x0000,
	}

	// CRC-16/X25标准配置
	CRC16_X25 = Config{
		Width:  16,
		Poly:   0x1021,
		Init:   0xFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFF,
	}

	// CRC-16/XMODEM标准配置
	CRC16_XMODEM = Config{
		Width:  16,
		Poly:   0xA001,
		Init:   0x00,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
	}

	// CRC-16/DNP标准配置
	CRC16_DNP = Config{
		Width:  16,
		Poly:   0x3D65,
		Init:   0x0000,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFF,
	}

	// CRC-16/ANSI标准配置
	CRC16_ANSI = Config{
		Width:  16,
		Poly:   0xA001,
		Init:   0x0000,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x0000,
	}

	// CRC-16/CCITT-Kermit标准配置
	CRC16_CCITT_KERMIT = Config{
		Width:  16,
		Poly:   0x1021,
		Init:   0x0000,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000,
	}

	// CRC-16/GENERIC标准配置
	CRC16_GENERIC = Config{
		Width:  16,
		Poly:   0xA001,
		Init:   0x0000,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x0000,
	}

	// CRC-16/CCITT-TRUE标准配置
	CRC16_CCITT_TRUE = Config{
		Width:  16,
		Poly:   0x1021,
		Init:   0x0000,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000,
	}

	// CRC-24/OPENPGP标准配置
	CRC24_OPENPGP = Config{
		Width:  24,
		Poly:   0x864CFB,
		Init:   0xB704CE,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x000000,
	}

	// CRC-32标准配置
	CRC32 = Config{
		Width:  32,
		Poly:   0x04C11DB7,
		Init:   0xFFFFFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFFFFFF,
	}

	// CRC-32/MPEG-2标准配置
	CRC32_MPEG2 = Config{
		Width:  32,
		Poly:   0x04C11DB7,
		Init:   0xFFFFFFFF,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00000000,
	}

	// CRC-32/PKZIP标准配置
	CRC32_PKZIP = Config{
		Width:  32,
		Poly:   0xEDB88320,
		Init:   0xFFFFFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFFFFFF,
	}

	// CRC-32C标准配置
	CRC32C = Config{
		Width:  32,
		Poly:   0x82F63B78,
		Init:   0xFFFFFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFFFFFF,
	}

	// CRC-32/CASTAGNOLI标准配置
	CRC32_CASTAGNOLI = Config{
		Width:  32,
		Poly:   0x1EDC6F41,
		Init:   0xFFFFFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0xFFFFFFFF,
	}

	// CRC-32/ADLER32标准配置
	CRC32_ADLER32 = Config{
		Width:  32,
		Poly:   0xFFFFFFFF,
		Init:   0x01,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00000000,
	}

	// CRC-64/ECMA标准配置
	CRC64_ECMA = Config{
		Width:  64,
		Poly:   0x42F0E1EBA9EA3693,
		Init:   0x0000000000000000,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000000000000000,
	}

	// CRC-64/ISO标准配置
	CRC64_ISO = Config{
		Width:  64,
		Poly:   0x42F0E1EBA9EA3693,
		Init:   0xFFFFFFFFFFFFFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000000000000000,
	}

	// CRC-64/WEIERSTRASS标准配置
	CRC64_WEIERSTRASS = Config{
		Width:  64,
		Poly:   0x42F0E1EBA9EA3693,
		Init:   0xFFFFFFFFFFFFFFFF,
		RefIn:  true,
		RefOut: true,
		XorOut: 0x0000000000000000,
	}
)
