/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-12-09 12:15:51
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-05-28 17:15:25
 * @FilePath: \go-toolbox\pkg\imgix\base.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package imgix

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

// LoadImageFromURL 从给定的 URL 加载图片并返回 image.Image 对象
func LoadImageFromURL(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get image from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return decodeImage(resp.Body)
}

// LoadImageFromFile 从给定的文件路径加载图片并返回 image.Image 对象
func LoadImageFromFile(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	return decodeImage(file)
}

// decodeImage 解码给定的 io.Reader 并返回 image.Image 对象
func decodeImage(reader io.Reader) (image.Image, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

// ImageFormat 定义图像格式类型
type ImageFormat int

// 定义支持的图像格式常量
const (
	JPEG ImageFormat = iota
	JPG
	PNG
	GIF
)

// String 返回图像格式的字符串表示
func (f ImageFormat) String() string {
	switch f {
	case JPEG:
		return "jpeg"
	case PNG:
		return "png"
	case GIF:
		return "gif"
	default:
		return "unknown"
	}
}

// WriterImage 将图像写入输出
func WriterImage(img image.Image, quality int, format ImageFormat, output io.Writer) error {
	switch format {
	case JPEG, JPG:
		if quality < 1 || quality > 100 {
			return errors.New("quality must be between 1 and 100 for JPEG")
		}
		return jpeg.Encode(output, img, &jpeg.Options{Quality: quality})
	case PNG:
		return png.Encode(output, img)
	case GIF:
		return gif.Encode(output, img, nil)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// EncodeImageToBase64 将图像编码为 Base64 字符串
func EncodeImageToBase64(img image.Image, format ImageFormat, quality int) (string, error) {
	var buf bytes.Buffer
	if err := WriterImage(img, quality, format, &buf); err != nil {
		return "", fmt.Errorf("failed to encode image to buffer: %w", err)
	}
	return "data:image/" + format.String() + ";base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// LoadImageFromFileAndEncodeToBase64 从文件加载图片并返回 Base64 编码字符串
func LoadImageFromFileAndEncodeToBase64(filePath string, format ImageFormat, quality int) (string, error) {
	img, err := LoadImageFromFile(filePath)
	if err != nil {
		return "", err
	}
	return EncodeImageToBase64(img, format, quality)
}

// WriterImageFromBytes 将字节切片转换为图像并写入输出
func WriterImageFromBytes(data []byte, quality int, format ImageFormat, output io.Writer) error {
	// 解码字节切片为图像
	img, err := decodeImage(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image from bytes: %w", err)
	}

	// 使用 WriterImage 函数将图像写入输出
	return WriterImage(img, quality, format, output)
}

// SaveBufToImageFile 将Buf数据保存到本地文件
func SaveBufToImageFile(buf *bytes.Buffer, filePath string, format ImageFormat) error {
	img, err := decodeImage(buf)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return WriterImage(img, 100, format, outFile) // 默认质量为 100
}

// AddOverlay 接受一个 img.Image 和一个颜色，返回添加了蒙层的 img.Image
func AddOverlay(originalImg image.Image, overlayColor color.RGBA) image.Image {
	// 获取原始图像的边界
	bounds := originalImg.Bounds()

	// 创建一个新的图像用于绘制蒙层
	maskedImg := image.NewRGBA(bounds)

	// 将原始图像绘制到新图像上
	draw.Draw(maskedImg, bounds, originalImg, image.Point{}, draw.Src)

	// 创建蒙层
	overlay := image.NewUniform(overlayColor)

	// 在原始图像上绘制蒙层
	draw.Draw(maskedImg, bounds, overlay, image.Point{}, draw.Over)

	return maskedImg
}

var DefaultImgPatterns = []string{"*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp", "*.webp", "*.tiff"}

// MatchesAnyPattern 判断 input 字符串是否匹配 patterns 中任意一个通配符模式（不区分大小写）
func MatchesAnyPattern(input string, patterns []string) bool {
	inputLower := strings.ToLower(input)
	for _, pattern := range patterns {
		patternLower := strings.ToLower(pattern)
		if strings.HasPrefix(patternLower, "*.") {
			ext := patternLower[1:]
			if strings.HasSuffix(inputLower, ext) {
				return true
			}
		}
	}
	return false
}
