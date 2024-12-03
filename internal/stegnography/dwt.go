package steganography

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type DWTSteganography struct{}

func NewDWTSteganography() *DWTSteganography {
	return &DWTSteganography{}
}

// Haar小波变换
func (d *DWTSteganography) dwt1D(data []float64) ([]float64, []float64) {
	n := len(data)
	approx := make([]float64, n/2)
	detail := make([]float64, n/2)

	for i := 0; i < n/2; i++ {
		approx[i] = (data[2*i] + data[2*i+1]) / math.Sqrt(2)
		detail[i] = (data[2*i] - data[2*i+1]) / math.Sqrt(2)
	}

	return approx, detail
}

// 逆Haar小波变换
func (d *DWTSteganography) idwt1D(approx, detail []float64) []float64 {
	n := len(approx) * 2
	result := make([]float64, n)

	for i := 0; i < len(approx); i++ {
		result[2*i] = (approx[i] + detail[i]) / math.Sqrt(2)
		result[2*i+1] = (approx[i] - detail[i]) / math.Sqrt(2)
	}

	return result
}

// 2D DWT变换
func (d *DWTSteganography) dwt2D(img [][]float64) ([][]float64, [][]float64, [][]float64, [][]float64) {
	rows := len(img)
	cols := len(img[0])

	// 水平方向变换
	tempH := make([][]float64, rows)
	for i := range tempH {
		tempH[i] = make([]float64, cols)
	}

	for i := 0; i < rows; i++ {
		approx, detail := d.dwt1D(img[i])
		copy(tempH[i][:cols/2], approx)
		copy(tempH[i][cols/2:], detail)
	}

	// 垂直方向变换
	ll := make([][]float64, rows/2)
	lh := make([][]float64, rows/2)
	hl := make([][]float64, rows/2)
	hh := make([][]float64, rows/2)

	for i := range ll {
		ll[i] = make([]float64, cols/2)
		lh[i] = make([]float64, cols/2)
		hl[i] = make([]float64, cols/2)
		hh[i] = make([]float64, cols/2)
	}

	for j := 0; j < cols; j++ {
		col := make([]float64, rows)
		for i := 0; i < rows; i++ {
			col[i] = tempH[i][j]
		}

		approx, detail := d.dwt1D(col)

		if j < cols/2 {
			for i := 0; i < rows/2; i++ {
				ll[i][j] = approx[i]
				hl[i][j] = detail[i]
			}
		} else {
			j2 := j - cols/2
			for i := 0; i < rows/2; i++ {
				lh[i][j2] = approx[i]
				hh[i][j2] = detail[i]
			}
		}
	}

	return ll, lh, hl, hh
}

func (d *DWTSteganography) EmbedText(img image.Image, text string) (image.Image, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// 确保图像尺寸是2的幂
	if !isPowerOfTwo(width) || !isPowerOfTwo(height) {
		return nil, fmt.Errorf("图像尺寸必须是2的幂")
	}

	// 将文本转换为比特流
	bits := textToBits(text)
	if len(bits) > (width*height)/64 { // 使用HL子带嵌入信息
		return nil, fmt.Errorf("文本太长，超出图像容量")
	}

	// 准备图像数据
	imgData := make([][]float64, height)
	for i := range imgData {
		imgData[i] = make([]float64, width)
		for j := 0; j < width; j++ {
			r, _, _, _ := img.At(j, i).RGBA()
			imgData[i][j] = float64(r >> 8)
		}
	}

	// DWT变换
	ll, lh, hl, hh := d.dwt2D(imgData)

	// 在HL子带中嵌入信息
	bitIndex := 0
	for i := 0; i < len(hl) && bitIndex < len(bits); i++ {
		for j := 0; j < len(hl[0]) && bitIndex < len(bits); j++ {
			if bits[bitIndex] == 1 {
				hl[i][j] = math.Abs(hl[i][j]) + 20
			} else {
				hl[i][j] = -math.Abs(hl[i][j]) - 20
			}
			bitIndex++
		}
	}

	// 逆变换
	result := d.idwt2D(ll, lh, hl, hh)

	// 创建结果图像
	outputImg := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := uint8(math.Max(0, math.Min(255, result[y][x])))
			outputImg.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}

	return outputImg, nil
}

func (d *DWTSteganography) ExtractText(img image.Image) (string, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// 准备图像数据
	imgData := make([][]float64, height)
	for i := range imgData {
		imgData[i] = make([]float64, width)
		for j := 0; j < width; j++ {
			r, _, _, _ := img.At(j, i).RGBA()
			imgData[i][j] = float64(r >> 8)
		}
	}

	// DWT变换
	_, _, hl, _ := d.dwt2D(imgData)

	// 从HL子带提取信息
	var bits []int
	for i := 0; i < len(hl); i++ {
		for j := 0; j < len(hl[0]); j++ {
			if hl[i][j] > 0 {
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

// 辅助函数
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

func textToBits(text string) []int {
	var bits []int
	for _, b := range []byte(text) {
		for i := 7; i >= 0; i-- {
			bits = append(bits, int((b>>uint(i))&1))
		}
	}
	// 添加结束标记
	for i := 7; i >= 0; i-- {
		bits = append(bits, 0)
	}
	return bits
}

func bitsToText(bits []int) string {
	var bytes []byte
	for i := 0; i < len(bits); i += 8 {
		if i+8 > len(bits) {
			break
		}
		var b byte
		for j := 0; j < 8; j++ {
			b = b<<1 | byte(bits[i+j])
		}
		bytes = append(bytes, b)
	}
	return string(bytes)
}

// 2D 逆DWT变换
func (d *DWTSteganography) idwt2D(ll, lh, hl, hh [][]float64) [][]float64 {
	rows := len(ll) * 2
	cols := len(ll[0]) * 2

	// 创建结果数组
	result := make([][]float64, rows)
	for i := range result {
		result[i] = make([]float64, cols)
	}

	// 临时数组用于存储水平方向的逆变换结果
	tempH := make([][]float64, rows)
	for i := range tempH {
		tempH[i] = make([]float64, cols)
	}

	// 垂直方向逆变换
	for j := 0; j < cols/2; j++ {
		// 处理左半部分
		colL := make([]float64, rows/2)
		colH := make([]float64, rows/2)
		for i := 0; i < rows/2; i++ {
			colL[i] = ll[i][j]
			colH[i] = hl[i][j]
		}
		col := d.idwt1D(colL, colH)
		for i := 0; i < rows; i++ {
			tempH[i][j] = col[i]
		}

		// 处理右半部分
		colL = make([]float64, rows/2)
		colH = make([]float64, rows/2)
		for i := 0; i < rows/2; i++ {
			colL[i] = lh[i][j]
			colH[i] = hh[i][j]
		}
		col = d.idwt1D(colL, colH)
		for i := 0; i < rows; i++ {
			tempH[i][j+cols/2] = col[i]
		}
	}

	// 水平方向逆变换
	for i := 0; i < rows; i++ {
		row := d.idwt1D(tempH[i][:cols/2], tempH[i][cols/2:])
		copy(result[i], row)
	}

	return result
}
