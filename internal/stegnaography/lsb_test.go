package steganography

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestLSB_EmbedAndExtract(t *testing.T) {
	// 创建一个测试用的 LSB 实例
	lsb := NewLSB()

	// 创建测试用例
	testCases := []struct {
		name    string
		text    string
		imgSize image.Rectangle
		wantErr bool
	}{
		{
			name:    "基本ASCII文本",
			text:    "Hello, World!",
			imgSize: image.Rect(0, 0, 100, 100),
			wantErr: false,
		},
		{
			name:    "中文文本",
			text:    "你好，世界！",
			imgSize: image.Rect(0, 0, 100, 100),
			wantErr: false,
		},
		{
			name:    "空文本",
			text:    "",
			imgSize: image.Rect(0, 0, 100, 100),
			wantErr: false,
		},
		{
			name:    "特殊字符",
			text:    "!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`",
			imgSize: image.Rect(0, 0, 100, 100),
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试图片
			img := image.NewRGBA(tc.imgSize)
			// 填充一些随机颜色
			for y := 0; y < tc.imgSize.Dy(); y++ {
				for x := 0; x < tc.imgSize.Dx(); x++ {
					img.Set(x, y, color.RGBA{
						R: uint8(x % 256),
						G: uint8(y % 256),
						B: uint8((x + y) % 256),
						A: 255,
					})
				}
			}

			// 嵌入文本
			encodedImg, err := lsb.EmbedText(img, tc.text)
			if (err != nil) != tc.wantErr {
				t.Errorf("EmbedText() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if encodedImg == nil && !tc.wantErr {
				t.Error("EmbedText() returned nil image but no error was expected")
				return
			}

			// 提取文本
			extractedText, err := lsb.ExtractText(encodedImg)
			if (err != nil) != tc.wantErr {
				t.Errorf("ExtractText() error = %v, wantErr %v", err, tc.wantErr)
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

// 添加一个简单的测试辅助函数
func TestLSB_Simple(t *testing.T) {
	lsb := NewLSB()
	text := "测试文本 Test Text 123"

	// 创建一个简单的测试图片
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// 填充白色背景
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}

	// 嵌入文本
	encodedImg, err := lsb.EmbedText(img, text)
	if err != nil {
		t.Fatalf("Failed to embed text: %v", err)
	}

	// 立即提取文本
	extractedText, err := lsb.ExtractText(encodedImg)
	if err != nil {
		t.Fatalf("Failed to extract text: %v", err)
	}

	// 比较结果
	if extractedText != text {
		t.Errorf("\nExpected: %q\nGot: %q", text, extractedText)
		t.Logf("Original length: %d, Extracted length: %d", len(text), len(extractedText))
		t.Logf("Original bytes: % x", []byte(text))
		t.Logf("Extracted bytes: % x", []byte(extractedText))
	} else {
		t.Logf("Success! Text was correctly embedded and extracted")
	}
}

func TestLSB_WithRealImage(t *testing.T) {
	lsb := NewLSB()
	testText := "Hello World! 你好，世界！@#$%^&*"

	// 读取测试图片
	imgFile, err := os.Open("Test.png")
	if err != nil {
		t.Fatalf("无法打开测试图片: %v", err)
	}
	defer imgFile.Close()

	// 解码原始图片
	originalImg, err := png.Decode(imgFile)
	if err != nil {
		t.Fatalf("无法解码测试图片: %v", err)
	}

	// 嵌入文本
	encodedImg, err := lsb.EmbedText(originalImg, testText)
	if err != nil {
		t.Fatalf("文本嵌入失败: %v", err)
	}

	// 保存处理后的图片
	outputFile, err := os.Create("Test_encoded.png")
	if err != nil {
		t.Fatalf("无法创建输出文件: %v", err)
	}

	// 使用png编码器的默认设置
	encoder := png.Encoder{
		CompressionLevel: png.DefaultCompression,
	}

	// 编码并保存图片
	if err := encoder.Encode(outputFile, encodedImg); err != nil {
		outputFile.Close()
		t.Fatalf("无法保存编码后的图片: %v", err)
	}
	outputFile.Close()

	// 验证保存的图片
	savedImg, err := os.Open("Test_encoded.png")
	if err != nil {
		t.Fatalf("无法打开保存的图片: %v", err)
	}
	defer savedImg.Close()

	// 解码保存的图片
	decodedImg, err := png.Decode(savedImg)
	if err != nil {
		t.Fatalf("无法解码保存的图片: %v", err)
	}

	// 提取文本
	extractedText, err := lsb.ExtractText(decodedImg)
	if err != nil {
		t.Fatalf("文本提取失败: %v", err)
	}

	// 验证提取的文本
	if extractedText != testText {
		t.Errorf("文本不匹配\n期望: %q\n实际: %q", testText, extractedText)
	} else {
		t.Logf("测试成功！文本正确嵌入并提取: %s", extractedText)
	}
}
