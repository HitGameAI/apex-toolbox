/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-09 17:15:16
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-11 10:33:31
 * @FilePath: \go-toolbox\pkg\crc\factory.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package crc

import (
	"fmt"
	"sync"
)

// Factory 定义计算器工厂接口
type Factory interface {
	Create() (Calculator, error) // 创建计算器实例，返回错误
}

// factoryImpl 是基础工厂的实现
type factoryImpl struct {
	cfg Config // CRC算法配置
}

// NewFactory 创建标准工厂
// 参数: cfg - CRC配置
// 返回: 工厂实例
func NewFactory(cfg Config) (Factory, error) {
	if cfg.Width == 0 || cfg.Width > 64 {
		return nil, fmt.Errorf("invalid width: %d, must be between 1 and 64", cfg.Width)
	}
	if cfg.Poly == 0 {
		return nil, fmt.Errorf("invalid polynomial: must be non-zero")
	}
	return &factoryImpl{cfg: cfg}, nil
}

// Create 创建计算器实例
func (f *factoryImpl) Create() (Calculator, error) {
	return New(f.cfg)
}

// cachedFactory 是带缓存的工厂实现
type cachedFactory struct {
	cfg      Config     // CRC算法配置
	instance Calculator // 缓存的计算器实例
	once     sync.Once  // 确保实例只创建一次
}

// NewCachedFactory 创建带缓存的工厂
// 参数: cfg - CRC配置
// 返回: 工厂实例 (线程安全)
func NewCachedFactory(cfg Config) Factory {
	return &cachedFactory{cfg: cfg}
}

// Create 创建计算器实例 (使用缓存)
func (f *cachedFactory) Create() (Calculator, error) {
	var err error
	f.once.Do(func() {
		f.instance, err = New(f.cfg) // 创建并缓存实例
	})
	return f.instance, err // 返回实例和错误
}

// 预定义标准工厂
var (
	// CRC-4/ITU标准配置工厂
	CRC4_ITUFactory = NewCachedFactory(CRC4_ITU)
	// CRC-5/EPC标准配置工厂
	CRC5_EPCFactory = NewCachedFactory(CRC5_EPC)
	// CRC-5/ITU标准配置工厂
	CRC5_ITUFactory = NewCachedFactory(CRC5_ITU)
	// CRC-5/USB标准配置工厂
	CRC5_USBFactory = NewCachedFactory(CRC5_USB)
	// CRC-6/ITU标准配置工厂
	CRC6_ITUFactory = NewCachedFactory(CRC6_ITU)
	// CRC-7/MMC标准配置工厂
	CRC7_MMCFactory = NewCachedFactory(CRC7_MMC)
	// CRC-8标准配置工厂
	CRC8Factory = NewCachedFactory(CRC8)
	// CRC-8/ATM标准配置工厂
	CRC8_ATMFactory = NewCachedFactory(CRC8_ATM)
	// CRC-8/CDMA2000标准配置工厂
	CRC8_CDMA2000Factory = NewCachedFactory(CRC8_CDMA2000)
	// CRC-8/DALLAS/1-WIRE标准配置工厂
	CRC8_DALLAS_1WIREFactory = NewCachedFactory(CRC8_DALLAS_1WIRE)
	// CRC-8/ITU标准配置工厂
	CRC8_ITUFactory = NewCachedFactory(CRC8_ITU)
	// CRC-8/ROHC标准配置工厂
	CRC8_ROHCFactory = NewCachedFactory(CRC8_ROHC)
	// CRC-8/MAXIM标准配置工厂
	CRC8_MAXIMFactory = NewCachedFactory(CRC8_MAXIM)
	// CRC-16/IBM标准配置工厂
	CRC16_IBMFactory = NewCachedFactory(CRC16_IBM)
	// CRC-16/MAXIM标准配置工厂
	CRC16_MAXIMFactory = NewCachedFactory(CRC16_MAXIM)
	// CRC-16/USB标准配置工厂
	CRC16_USBFactory = NewCachedFactory(CRC16_USB)
	// CRC-16/MODBUS标准配置工厂
	CRC16_MODBUSFactory = NewCachedFactory(CRC16_MODBUS)
	// CRC-16/DNP标准配置工厂
	CRC16_DNPFactory = NewCachedFactory(CRC16_DNP)
	// CRC-16/ANSI标准配置工厂
	CRC16_ANSIFactory = NewCachedFactory(CRC16_ANSI)
	// CRC-16/XMODEM标准配置工厂
	CRC16_XMODEMFactory = NewCachedFactory(CRC16_XMODEM)
	// CRC-16/CCITT标准配置工厂
	CRC16_CCITTFactory = NewCachedFactory(CRC16_CCITT)
	// CRC-16/CCITT-FALSE标准配置工厂
	CRC16_CCITT_FALSEFactory = NewCachedFactory(CRC16_CCITT_FALSE)
	// CRC-16/CCITT-Kermit标准配置工厂
	CRC16_CCITT_KERMITFactory = NewCachedFactory(CRC16_CCITT_KERMIT)
	// CRC-16/X25标准配置工厂
	CRC16_X25Factory = NewCachedFactory(CRC16_X25)
	// CRC-32标准配置工厂
	CRC32Factory = NewCachedFactory(CRC32)
	// CRC-32/MPEG-2标准配置工厂
	CRC32_MPEG2Factory = NewCachedFactory(CRC32_MPEG2)
	// CRC-32/PKZIP标准配置工厂
	CRC32_PKZIPFactory = NewCachedFactory(CRC32_PKZIP)
	// CRC-32C标准配置工厂
	CRC32CFactory = NewCachedFactory(CRC32C)
	// CRC-24/OPENPGP标准配置工厂
	CRC24_OPENPGPFactory = NewCachedFactory(CRC24_OPENPGP)
	// CRC-64/ECMA标准配置工厂
	CRC64_ECMAFactory = NewCachedFactory(CRC64_ECMA)
	// CRC-64/ISO标准配置工厂
	CRC64_ISOFactory = NewCachedFactory(CRC64_ISO)
	// CRC-64/WEIERSTRASS标准配置工厂
	CRC64_WEIERSTRASSFactory = NewCachedFactory(CRC64_WEIERSTRASS)
	// CRC-32/CASTAGNOLI标准配置工厂
	CRC32_CASTAGNOLIFactory = NewCachedFactory(CRC32_CASTAGNOLI)
	// CRC-16/GENERIC标准配置工厂
	CRC16_GENERICFactory = NewCachedFactory(CRC16_GENERIC)
	// CRC-16/CCITT-TRUE标准配置工厂
	CRC16_CCITT_TRUEFactory = NewCachedFactory(CRC16_CCITT_TRUE)
	// CRC-32/ADLER32标准配置工厂
	CRC32_ADLER32Factory = NewCachedFactory(CRC32_ADLER32)
)
