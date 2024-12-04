package steganography

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
)

func TestDWTSteganography_EmbedAndExtract(t *testing.T) {
	dwt := NewDWTSteganography()

	// 创建测试用例
	testCases := []struct {
		name    string
		text    string
		imgSize int
		wantErr bool
	}{
		{
			name:    "基本ASCII文本",
			text:    "Hello, World!",
			imgSize: 256, // 必须是2的幂
			wantErr: false,
		},
		{
			name:    "中文文本",
			text:    "你好，世界！",
			imgSize: 256,
			wantErr: false,
		},
		{
			name:    "空文本",
			text:    "",
			imgSize: 256,
			wantErr: false,
		},
		{
			name:    "特殊字符",
			text:    "!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`",
			imgSize: 256,
			wantErr: false,
		},
		{
			name:    "非法图像尺寸",
			text:    "test",
			imgSize: 100, // 不是2的幂
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试图像
			img := image.NewRGBA(image.Rect(0, 0, tc.imgSize, tc.imgSize))
			// 填充随机颜色
			for y := 0; y < tc.imgSize; y++ {
				for x := 0; x < tc.imgSize; x++ {
					img.Set(x, y, color.RGBA{
						R: uint8(x % 256),
						G: uint8(y % 256),
						B: uint8((x + y) % 256),
						A: 255,
					})
				}
			}

			// 嵌入文本
			encodedImg, err := dwt.EmbedText(img, tc.text)
			if (err != nil) != tc.wantErr {
				t.Errorf("EmbedText() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr {
				return
			}

			// 提取文本
			extractedText, err := dwt.ExtractText(encodedImg)
			if err != nil {
				t.Errorf("ExtractText() error = %v", err)
				return
			}

			// 比较原始文本和提取的文本
			if extractedText != tc.text {
				t.Errorf("Text mismatch:\nwant: %q\ngot:  %q", tc.text, extractedText)
				// 打印二进制比较以帮助调试
				t.Logf("Original text bytes: % x", []byte(tc.text))
				t.Logf("Extracted text bytes: % x", []byte(extractedText))
			}
		})
	}
}

func TestDWT_Transform(t *testing.T) {
	dwt := NewDWTSteganography()

	// 测试1D DWT变换
	t.Run("1D DWT Transform", func(t *testing.T) {
		data := []float64{1, 2, 3, 4, 5, 6, 7, 8}
		approx, detail := dwt.dwt1D(data)
		recovered := dwt.idwt1D(approx, detail)

		// 检查恢复的数据
		for i, val := range data {
			if math.Abs(val-recovered[i]) > 0.1 {
				t.Errorf("1D DWT transform error at [%d]: original=%f, recovered=%f",
					i, val, recovered[i])
			}
		}
	})

	// 测试2D DWT变换
	t.Run("2D DWT Transform", func(t *testing.T) {
		// 创建测试数据
		size := 8
		data := make([][]float64, size)
		for i := range data {
			data[i] = make([]float64, size)
			for j := range data[i] {
				data[i][j] = float64((i + j) % 256)
			}
		}

		// 执行2D DWT变换
		ll, lh, hl, hh := dwt.dwt2D(data)

		// 检查子带大小
		expectedSize := size / 2
		if len(ll) != expectedSize || len(lh) != expectedSize ||
			len(hl) != expectedSize || len(hh) != expectedSize {
			t.Errorf("Incorrect subband size: expected %d, got %d, %d, %d, %d",
				expectedSize, len(ll), len(lh), len(hl), len(hh))
		}

		// 检查子带的值范围
		checkSubband := func(name string, subband [][]float64) {
			for i := range subband {
				for j := range subband[i] {
					if math.IsNaN(subband[i][j]) || math.IsInf(subband[i][j], 0) {
						t.Errorf("Invalid value in %s at [%d][%d]: %f",
							name, i, j, subband[i][j])
					}
				}
			}
		}

		checkSubband("LL", ll)
		checkSubband("LH", lh)
		checkSubband("HL", hl)
		checkSubband("HH", hh)
	})
}

func TestIsPowerOfTwo(t *testing.T) {
	testCases := []struct {
		input    int
		expected bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, false},
		{4, true},
		{8, true},
		{16, true},
		{100, false},
		{256, true},
		{1024, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input_%d", tc.input), func(t *testing.T) {
			result := isPowerOfTwo(tc.input)
			if result != tc.expected {
				t.Errorf("isPowerOfTwo(%d) = %v; want %v",
					tc.input, result, tc.expected)
			}
		})
	}
}
