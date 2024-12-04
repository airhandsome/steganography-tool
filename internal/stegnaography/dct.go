package steganography

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type DCTSteganography struct {
	blockSize int
}

func NewDCTSteganography() *DCTSteganography {
	return &DCTSteganography{
		blockSize: 8,
	}
}

func (d *DCTSteganography) EmbedText(img image.Image, text string) (image.Image, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// 检查图像尺寸
	if width%d.blockSize != 0 || height%d.blockSize != 0 {
		return nil, fmt.Errorf("图像尺寸必须是%d的倍数", d.blockSize)
	}

	// 将文本转换为比特流
	bits := textToBits(text)
	maxBits := (width * height) / (d.blockSize * d.blockSize)
	if len(bits) > maxBits {
		return nil, fmt.Errorf("文本太长，超出图像容量")
	}

	// 创建输出图像
	output := image.NewRGBA(bounds)
	bitIndex := 0

	// 按块处理图像
	for y := 0; y < height; y += d.blockSize {
		for x := 0; x < width; x += d.blockSize {
			if bitIndex >= len(bits) {
				// 复制剩余的图像块
				d.copyBlock(img, output, x, y)
				continue
			}

			// 提取块数据
			block := d.getBlock(img, x, y)
			// 预处理
			block = d.preprocessBlock(block)
			// DCT变换
			dctBlock := d.dct2D(block)

			// 在中频系数中嵌入信息
			if bits[bitIndex] == 1 {
				dctBlock[4][3] = math.Abs(dctBlock[4][3]) + 25.0
			} else {
				dctBlock[4][3] = -math.Abs(dctBlock[4][3]) - 25.0
			}
			bitIndex++

			// 逆DCT变换
			idctBlock := d.idct2D(dctBlock)
			// 后处理
			idctBlock = d.postprocessBlock(idctBlock)
			// 写回图像
			d.setBlock(output, idctBlock, x, y)
		}
	}

	return output, nil
}

func (d *DCTSteganography) ExtractText(img image.Image) (string, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	var bits []int

	// 按块处理图像
	for y := 0; y < height; y += d.blockSize {
		for x := 0; x < width; x += d.blockSize {
			// 提取块数据
			block := d.getBlock(img, x, y)

			// DCT变换
			dctBlock := d.dct2D(block)

			// 从中频系数提取信息
			if dctBlock[4][3] > 0 {
				bits = append(bits, 1)
			} else {
				bits = append(bits, 0)
			}

			// 检查结束标记
			if len(bits)%8 == 0 {
				text := bitsToText(bits)
				if len(text) > 0 && text[len(text)-1] == 0 {
					return text[:len(text)-1], nil
				}
			}
		}
	}

	return bitsToText(bits), nil
}

// 辅助方法：获取图像块
func (d *DCTSteganography) getBlock(img image.Image, x, y int) [][]float64 {
	block := make([][]float64, d.blockSize)
	for i := range block {
		block[i] = make([]float64, d.blockSize)
		for j := 0; j < d.blockSize; j++ {
			r, _, _, _ := img.At(x+j, y+i).RGBA()
			block[i][j] = float64(r >> 8)
		}
	}
	return block
}

// 辅助方法：设置图像块
func (d *DCTSteganography) setBlock(img *image.RGBA, block [][]float64, x, y int) {
	for i := 0; i < d.blockSize; i++ {
		for j := 0; j < d.blockSize; j++ {
			val := uint8(math.Max(0, math.Min(255, block[i][j])))
			img.Set(x+j, y+i, color.RGBA{val, val, val, 255})
		}
	}
}

// 辅助方法：复制图像块
func (d *DCTSteganography) copyBlock(src image.Image, dst *image.RGBA, x, y int) {
	for i := 0; i < d.blockSize; i++ {
		for j := 0; j < d.blockSize; j++ {
			dst.Set(x+j, y+i, src.At(x+j, y+i))
		}
	}
}

// 2D DCT变换
func (d *DCTSteganography) dct2D(block [][]float64) [][]float64 {
	n := d.blockSize
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
	}

	// 计算DCT系数
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			var sum float64
			for x := 0; x < n; x++ {
				for y := 0; y < n; y++ {
					// DCT变换公式
					cos1 := math.Cos((2*float64(x) + 1) * float64(u) * math.Pi / (2 * float64(n)))
					cos2 := math.Cos((2*float64(y) + 1) * float64(v) * math.Pi / (2 * float64(n)))
					sum += block[x][y] * cos1 * cos2
				}
			}

			// 计算系数
			cu := 1.0
			cv := 1.0
			if u == 0 {
				cu = 1.0 / math.Sqrt(2)
			}
			if v == 0 {
				cv = 1.0 / math.Sqrt(2)
			}

			// 保存结果
			result[u][v] = (2 * cu * cv * sum) / float64(n)
		}
	}

	return result
}

// 2D 逆DCT变换
func (d *DCTSteganography) idct2D(block [][]float64) [][]float64 {
	n := d.blockSize
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
	}

	// 计算逆DCT
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			var sum float64
			for u := 0; u < n; u++ {
				for v := 0; v < n; v++ {
					// 计算系数
					cu := 1.0
					cv := 1.0
					if u == 0 {
						cu = 1.0 / math.Sqrt(2)
					}
					if v == 0 {
						cv = 1.0 / math.Sqrt(2)
					}

					// 逆DCT变换公式
					cos1 := math.Cos((2*float64(x) + 1) * float64(u) * math.Pi / (2 * float64(n)))
					cos2 := math.Cos((2*float64(y) + 1) * float64(v) * math.Pi / (2 * float64(n)))
					sum += cu * cv * block[u][v] * cos1 * cos2
				}
			}

			// 保存结果
			result[x][y] = (2 * sum) / float64(n)
		}
	}

	return result
}

// 添加一个辅助方法来预处理图像块
func (d *DCTSteganography) preprocessBlock(block [][]float64) [][]float64 {
	n := d.blockSize
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			// 将像素值从[0,255]映射到[-128,127]
			result[i][j] = block[i][j] - 128.0
		}
	}
	return result
}

// 添加一个辅助方法来后处理图像块
func (d *DCTSteganography) postprocessBlock(block [][]float64) [][]float64 {
	n := d.blockSize
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			// 将值映射回[0,255]范围
			result[i][j] = block[i][j] + 128.0
			// 确保值在有效范围内
			result[i][j] = math.Max(0, math.Min(255, result[i][j]))
		}
	}
	return result
}
