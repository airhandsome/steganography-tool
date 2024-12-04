package steganography

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestDCTSteganography_EmbedAndExtract(t *testing.T) {
	dct := NewDCTSteganography()

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
			imgSize: 128, // 必须是8的倍数
			wantErr: false,
		},
		{
			name:    "中文文本",
			text:    "你好，世界！",
			imgSize: 128,
			wantErr: false,
		},
		{
			name:    "空文本",
			text:    "",
			imgSize: 128,
			wantErr: false,
		},
		{
			name:    "特殊字符",
			text:    "!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`",
			imgSize: 128,
			wantErr: false,
		},
		{
			name:    "非法图像尺寸",
			text:    "test",
			imgSize: 30, // 不是8的倍数
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
			encodedImg, err := dct.EmbedText(img, tc.text)
			if (err != nil) != tc.wantErr {
				t.Errorf("EmbedText() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr {
				return
			}

			// 提取文本
			extractedText, err := dct.ExtractText(encodedImg)
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

func TestDCT_Transform(t *testing.T) {
	dct := NewDCTSteganography()

	// 测试DCT变换和逆变换
	block := make([][]float64, 8)
	for i := range block {
		block[i] = make([]float64, 8)
		for j := range block[i] {
			block[i][j] = float64((i + j) % 256)
		}
	}

	// 执行DCT变换
	dctBlock := dct.dct2D(block)
	// 执行逆DCT变换
	idctBlock := dct.idct2D(dctBlock)

	// 检查变换后的数据是否接近原始数据
	for i := range block {
		for j := range block[i] {
			diff := math.Abs(block[i][j] - idctBlock[i][j])
			if diff > 0.1 { // 允许一定的误差
				t.Errorf("DCT transform error at [%d][%d]: original=%f, recovered=%f",
					i, j, block[i][j], idctBlock[i][j])
			}
		}
	}
}
