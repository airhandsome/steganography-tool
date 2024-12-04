package steganography

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
)

type LSB struct{}

func NewLSB() *LSB {
	return &LSB{}
}

func (l *LSB) EmbedText(img image.Image, text string) (*image.RGBA, error) {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	// 首先将原始图片复制到新的RGBA图片中
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	// 将文本转换为UTF-8字节数组
	textBytes := []byte(text)
	// 添加结束标记
	textBytes = append(textBytes, 0)

	// 将字节数组转换为二进制字符串
	var binText strings.Builder
	for _, b := range textBytes {
		fmt.Fprintf(&binText, "%08b", b)
	}

	binString := binText.String()
	binLen := len(binString)

	// 检查图片容量是否足够
	if binLen > (bounds.Dx() * bounds.Dy()) {
		return nil, fmt.Errorf("图片太小，无法存储这么多文本")
	}

	// 嵌入文本数据
	binIndex := 0
	for y := bounds.Min.Y; y < bounds.Max.Y && binIndex < binLen; y++ {
		for x := bounds.Min.X; x < bounds.Max.X && binIndex < binLen; x++ {
			// 获取当前像素的RGBA值
			r, g, b, a := rgba.At(x, y).RGBA()

			// 将颜色值转换为8位
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// 修改红色通道的最低位
			if binString[binIndex] == '1' {
				r8 |= 1 // 设置最低位为1
			} else {
				r8 &= 0xFE // 设置最低位为0
			}
			binIndex++

			// 设置修改后的像素值
			rgba.Set(x, y, color.RGBA{
				R: r8,
				G: g8,
				B: b8,
				A: a8,
			})
		}
	}

	return rgba, nil
}

func (l *LSB) ExtractText(img image.Image) (string, error) {
	bounds := img.Bounds()
	var result []byte
	bitCount := 0
	currentByte := uint8(0)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 获取红色通道值
			r, _, _, _ := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)

			// 提取最低位
			bit := r8 & 1

			// 构建字节
			currentByte = (currentByte << 1) | bit
			bitCount++

			if bitCount == 8 {
				// 检查是否到达结束标记
				if currentByte == 0 {
					return string(result), nil
				}

				result = append(result, currentByte)
				bitCount = 0
				currentByte = 0
			}
		}
	}

	return string(result), nil
}

// 辅助函数：将文本转换为二进制字符串
func textToBinary(text string) string {
	var binary strings.Builder
	for _, char := range text {
		// 使用 %08b 确保每个字符都是8位
		fmt.Fprintf(&binary, "%08b", char)
	}
	// 添加结束标记
	fmt.Fprintf(&binary, "%08b", 0)
	return binary.String()
}

// 辅助函数：将二进制字符串转换为文本
func binaryToText(binary string) string {
	var text strings.Builder
	for i := 0; i < len(binary); i += 8 {
		if i+8 > len(binary) {
			break
		}
		if n, err := strconv.ParseUint(binary[i:i+8], 2, 8); err == nil {
			// 检查是否到达结束标记
			if n == 0 {
				break
			}
			text.WriteRune(rune(n))
		}
	}
	return text.String()
}
