package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image"
	"image/color"
	"image/png"
	steganography "steganography-tool/internal/stegnaography"
)

type SteganoUI struct {
	window           fyne.Window
	lsb              *steganography.LSB
	dct              *steganography.DCTSteganography
	dwt              *steganography.DWTSteganography
	imageView        *canvas.Image
	textInput        *widget.Entry
	resultText       *widget.RichText
	algorithm        *widget.Select // 新增：算法选择下拉框
	textLength       *widget.Label  // 新增：文本长度显示
	currentImageSize image.Point    // 新增：存储当前图片尺寸
}

func NewSteganoUI(app fyne.App) *SteganoUI {
	icon, err := fyne.LoadResourceFromPath("assets/secret.png")
	if err == nil {
		app.SetIcon(icon)
	}
	app.Settings().SetTheme(NewCustomTheme())

	ui := &SteganoUI{
		window:     app.NewWindow("跟你说悄悄话"),
		lsb:        steganography.NewLSB(),
		dct:        steganography.NewDCTSteganography(),
		dwt:        steganography.NewDWTSteganography(),
		textInput:  widget.NewMultiLineEntry(),
		textLength: widget.NewLabel(""), // 初始化文本长度标签
	}

	// 初始化算法选择下拉框
	ui.algorithm = widget.NewSelect([]string{"LSB", "DCT", "DWT"}, func(value string) {
		// 当选择改变时更新文本长度显示
		ui.updateTextLength()
	})
	ui.algorithm.SetSelected("LSB") // 设置默认选项

	// 设置文本输入变化的回调
	ui.textInput.OnChanged = func(text string) {
		ui.updateTextLength()
	}

	ui.window.Resize(fyne.NewSize(800, 500))
	ui.window.CenterOnScreen()
	ui.window.SetIcon(icon)

	ui.createUI()

	return ui
}

// 新增：更新文本长度显示的方法
func (s *SteganoUI) updateTextLength() {
	text := s.textInput.Text
	algorithm := s.algorithm.Selected
	length := len(text)

	var maxLength int
	if s.imageView != nil && s.imageView.Image != nil {
		// 获取当前图片尺寸
		width := s.currentImageSize.X
		height := s.currentImageSize.Y

		// 计算最大容量
		maxLength = s.calculateMaxCapacity(width, height, algorithm)

		// 预留20%空间给元数据和结束标记
		maxLength = int(float64(maxLength) * 0.8)
	} else {
		maxLength = 0
	}

	// 更新显示
	s.textLength.SetText(fmt.Sprintf("文本长度: %d/%d 字节", length, maxLength))

	// 根据长度设置显示样式
	if length > maxLength {
		s.textLength.TextStyle = fyne.TextStyle{Bold: true}
	} else if float64(length) > float64(maxLength)*0.9 {
		s.textLength.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		s.textLength.TextStyle = fyne.TextStyle{}
	}
}

// 在 SteganoUI 结构体中添加一个方法来计算最大容量
func (s *SteganoUI) calculateMaxCapacity(width, height int, algorithm string) int {
	pixelCount := width * height

	switch algorithm {
	case "LSB":
		// LSB算法：每个像素3个通道，每个通道可以存储1位
		// 总比特数除以8得到字节数
		return (pixelCount * 3) / 8

	case "DCT":
		// DCT算法：基于8x8的块
		blocks := (width / 8) * (height / 8)
		// 每个块可以存储1位，总比特数除以8得到字节数
		return (blocks) / 8

	case "DWT":
		// DWT算法：使用高频分量存储数据
		// 假设使用HL子带，大小为原图的1/4
		return (pixelCount / 16) // 除以4获取子带大小，再除以4考虑实际可用容量

	default:
		return 0
	}
}

func (s *SteganoUI) createUI() {
	// 创建标签页
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("隐藏悄悄话", theme.ContentAddIcon(), s.createEncryptTab()),
		container.NewTabItemWithIcon("查看悄悄话", theme.ContentClearIcon(), s.createDecryptTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	s.window.SetContent(tabs)
	s.window.CenterOnScreen()
}

// 添加图片预处理方法
func (s *SteganoUI) preprocessImage(img image.Image, algorithm string) (image.Image, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	switch algorithm {
	case "DCT":
		// DCT需要8的倍数
		newWidth := width - (width % 8)
		newHeight := height - (height % 8)

		if newWidth == 0 || newHeight == 0 {
			return nil, fmt.Errorf("图片尺寸太小，需要至少8x8像素")
		}

		// 创建新图像
		newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		for y := 0; y < newHeight; y++ {
			for x := 0; x < newWidth; x++ {
				newImg.Set(x, y, img.At(x, y))
			}
		}
		return newImg, nil

	case "DWT":
		// DWT需要2的幂
		newWidth := nearestPowerOfTwo(width)
		newHeight := nearestPowerOfTwo(height)

		if newWidth == 0 || newHeight == 0 {
			return nil, fmt.Errorf("图片尺寸太小，需要至少2x2像素")
		}

		// 创建新图像
		newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		for y := 0; y < newHeight; y++ {
			for x := 0; x < newWidth; x++ {
				newImg.Set(x, y, img.At(x, y))
			}
		}
		return newImg, nil

	default:
		// LSB不需要特殊处理
		return img, nil
	}
}

// 辅助函数：找到最近的2的幂（向下取整）
func nearestPowerOfTwo(n int) int {
	power := 1
	for power*2 <= n {
		power *= 2
	}
	return power
}

func (s *SteganoUI) createEncryptTab() fyne.CanvasObject {
	// 创建图片容器
	imageContainer := container.NewVBox()
	// 创建图片预览区
	placeholder, err := fyne.LoadResourceFromPath("assets/placeholder.png")
	if err != nil {
		placeholder = theme.FyneLogo()
	}
	s.imageView = canvas.NewImageFromResource(placeholder) // 需要替换为默认图片
	s.imageView.SetMinSize(fyne.NewSize(350, 350))
	s.imageView.FillMode = canvas.ImageFillContain
	imageContainer.Add(s.imageView)
	// 创建图片上传区
	imageCard := widget.NewCard(
		"",
		"图片预览",
		container.NewVBox(
			imageContainer,
			widget.NewButtonWithIcon("选择图片", theme.FolderOpenIcon(), func() {
				fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						dialog.ShowError(err, s.window)
						return
					}
					if reader == nil {
						return
					}

					originalImg, _, err := image.Decode(reader)
					if err != nil {
						dialog.ShowError(fmt.Errorf("无法加载图片: %v", err), s.window)
						return
					}
					defer reader.Close()

					// 更新当前图片尺寸
					bounds := originalImg.Bounds()
					s.currentImageSize = image.Point{
						X: bounds.Dx(),
						Y: bounds.Dy(),
					}

					// 更新图片显示
					newImage := canvas.NewImageFromImage(originalImg)
					newImage.SetMinSize(fyne.NewSize(350, 350))
					newImage.FillMode = canvas.ImageFillContain
					imageContainer.Remove(s.imageView)
					s.imageView = newImage
					imageContainer.Add(s.imageView)
					imageContainer.Refresh()

					// 更新文本长度显示
					s.updateTextLength()
				}, s.window)
				fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
				fd.Show()
			}),
		),
	)

	// 创建算法选择区域
	algorithmCard := widget.NewCard(
		"",
		"选择算法",
		container.NewVBox(
			s.algorithm,
			s.textLength,
		),
	)

	// 创建文本输入区
	s.textInput.SetPlaceHolder("在此输入要隐藏的文本...")
	s.textInput.MultiLine = true
	s.textInput.Wrapping = fyne.TextWrapWord
	s.textInput.SetMinRowsVisible(10)

	textCard := widget.NewCard(
		"",
		"文本输入",
		container.NewVBox(
			s.textInput,
			widget.NewButtonWithIcon("加密并保存", theme.DocumentSaveIcon(), func() {
				if s.imageView.Image == nil {
					dialog.ShowError(fmt.Errorf("请先选择图片"), s.window)
					return
				}
				if s.textInput.Text == "" {
					dialog.ShowError(fmt.Errorf("请输入要隐藏的文本"), s.window)
					return
				}

				var encodedImg image.Image
				var err error

				// 预处理图片
				processedImg, err := s.preprocessImage(s.imageView.Image, s.algorithm.Selected)
				if err != nil {
					dialog.ShowError(err, s.window)
					return
				}

				// 如果图片尺寸发生变化，显示提示
				// if processedImg.Bounds().Dx() != s.imageView.Image.Bounds().Dx() ||
				// 	processedImg.Bounds().Dy() != s.imageView.Image.Bounds().Dy() {
				// 	dialog.ShowInformation("提示", fmt.Sprintf(
				// 		"图片已被裁剪至 %dx%d 以适应算法要求",
				// 		processedImg.Bounds().Dx(),
				// 		processedImg.Bounds().Dy(),
				// 	), s.window)
				// }

				// 根据选择的算法执行相应的嵌入操作
				switch s.algorithm.Selected {
				case "LSB":
					encodedImg, err = s.lsb.EmbedText(processedImg, s.textInput.Text)
				case "DCT":
					encodedImg, err = s.dct.EmbedText(processedImg, s.textInput.Text)
				case "DWT":
					encodedImg, err = s.dwt.EmbedText(processedImg, s.textInput.Text)
				}

				if err != nil {
					dialog.ShowError(fmt.Errorf("加密失败: %v", err), s.window)
					return
				}

				fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						dialog.ShowError(err, s.window)
						return
					}
					if writer == nil {
						return
					}
					defer writer.Close()

					err = png.Encode(writer, encodedImg)
					if err != nil {
						dialog.ShowError(fmt.Errorf("保存失败: %v", err), s.window)
						return
					}
					dialog.ShowInformation("成功", "图片已成功保存", s.window)
				}, s.window)
				fd.SetFileName("encoded_image.png")
				fd.Show()
			}),
		),
	)

	// 使用垂直布局组合算法选择和文本输入
	rightContainer := container.NewVBox(
		algorithmCard,
		textCard,
	)

	// 使用分割容器
	split := container.NewHSplit(imageCard, rightContainer)
	split.SetOffset(0.5)

	return split
}

func (s *SteganoUI) createDecryptTab() fyne.CanvasObject {
	// 创建图片容器
	imageContainer := container.NewVBox()
	// 创建图片预览区
	placeholder, err := fyne.LoadResourceFromPath("assets/placeholder.png")
	if err != nil {
		placeholder = theme.FyneLogo()
	}
	decryptImageView := canvas.NewImageFromResource(placeholder)
	decryptImageView.SetMinSize(fyne.NewSize(350, 350))
	decryptImageView.FillMode = canvas.ImageFillContain
	imageContainer.Add(decryptImageView)

	// 保存当前图片的变量
	var currentImg image.Image

	// 创建算法选择
	algorithmSelect := widget.NewSelect([]string{"LSB", "DCT", "DWT"}, func(selected string) {
		// 当算法改变时，如果已有图片，则重新解密
		if currentImg != nil {
			// 预处理图片
			processedImg, err := s.preprocessImage(currentImg, selected)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			// 使用新算法提取文本
			var text string
			switch selected {
			case "LSB":
				text, err = s.lsb.ExtractText(processedImg)
			case "DCT":
				text, err = s.dct.ExtractText(processedImg)
			case "DWT":
				text, err = s.dwt.ExtractText(processedImg)
			}

			if err != nil {
				dialog.ShowError(fmt.Errorf("解密失败: %v", err), s.window)
				return
			}

			// 更新文本显示
			s.resultText.Segments = []widget.RichTextSegment{
				&widget.TextSegment{
					Style: widget.RichTextStyle{
						SizeName:  theme.SizeNameText,
						ColorName: theme.ColorNameForeground,
						TextStyle: fyne.TextStyle{Bold: true},
					},
					Text: text,
				},
			}
			s.resultText.Refresh()
		}
	})
	algorithmSelect.SetSelected("LSB")

	// 创建图片上传区
	imageCard := widget.NewCard(
		"",
		"图片预览",
		container.NewVBox(
			imageContainer,
			container.NewHBox(
				widget.NewLabel("选择算法:"),
				algorithmSelect,
			),
			widget.NewButtonWithIcon("选择图片", theme.FolderOpenIcon(), func() {
				fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						dialog.ShowError(err, s.window)
						return
					}
					if reader == nil {
						return
					}

					img, _, err := image.Decode(reader)
					if err != nil {
						dialog.ShowError(fmt.Errorf("无法加载图片: %v", err), s.window)
						return
					}
					defer reader.Close()

					// 保存当前图片
					currentImg = img

					// 预处理图片
					processedImg, err := s.preprocessImage(img, algorithmSelect.Selected)
					if err != nil {
						dialog.ShowError(err, s.window)
						return
					}

					// 更新图片显示
					newImage := canvas.NewImageFromImage(processedImg)
					newImage.SetMinSize(fyne.NewSize(350, 350))
					newImage.FillMode = canvas.ImageFillContain
					imageContainer.Remove(decryptImageView)
					decryptImageView = newImage
					imageContainer.Add(decryptImageView)
					imageContainer.Refresh()

					// 提取文本
					var text string
					switch algorithmSelect.Selected {
					case "LSB":
						text, err = s.lsb.ExtractText(processedImg)
					case "DCT":
						text, err = s.dct.ExtractText(processedImg)
					case "DWT":
						text, err = s.dwt.ExtractText(processedImg)
					}

					if err != nil {
						dialog.ShowError(fmt.Errorf("解密失败: %v", err), s.window)
						return
					}

					s.resultText.Segments = []widget.RichTextSegment{
						&widget.TextSegment{
							Style: widget.RichTextStyle{
								SizeName:  theme.SizeNameText,
								ColorName: theme.ColorNameForeground,
								TextStyle: fyne.TextStyle{Bold: true},
							},
							Text: text,
						},
					}
					s.resultText.Refresh()
				}, s.window)
				fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
				fd.Show()
			}),
		),
	)

	// 创建文本显示区
	s.resultText = widget.NewRichText()
	s.resultText.Wrapping = fyne.TextWrapWord

	// 创建一个带有边框和背景的文本显示区
	textBorder := canvas.NewRectangle(color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	textBackground := canvas.NewRectangle(color.White)

	// 创建标题
	title := widget.NewLabelWithStyle("提取的文本：", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 创建滚动容器来包装文本显示
	scrollContainer := container.NewScroll(s.resultText)
	scrollContainer.SetMinSize(fyne.NewSize(0, 200)) // 设置固定高度

	// 创建复制按钮
	copyBtn := widget.NewButtonWithIcon("复制文本", theme.ContentCopyIcon(), func() {
		if len(s.resultText.Segments) > 0 {
			if textSegment, ok := s.resultText.Segments[0].(*widget.TextSegment); ok {
				s.window.Clipboard().SetContent(textSegment.Text)
				dialog.ShowInformation("成功", "文本已复制到剪贴板", s.window)
			}
		} else {
			dialog.ShowInformation("提示", "没有可复制的文本", s.window)
		}
	})

	textCard := widget.NewCard(
		"",
		"提取结果",
		container.NewVBox(
			container.NewMax(
				textBorder,
				container.NewPadded(
					container.NewMax(
						textBackground,
						container.NewVBox(
							title,
							container.NewPadded(scrollContainer), // 使用滚动容器
						),
					),
				),
			),
			copyBtn,
		),
	)

	// 使用分割容器
	split := container.NewHSplit(imageCard, textCard)
	split.SetOffset(0.5)

	return split
}

func (s *SteganoUI) ShowAndRun() {
	s.window.ShowAndRun()
}
