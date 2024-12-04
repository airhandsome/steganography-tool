@echo off
echo 正在清理依赖...
go mod tidy

echo 正在安装打包工具...
go install fyne.io/fyne/v2/cmd/fyne@latest

echo 正在打包应用...
fyne package -os windows -icon .\assets\secret.png -name "silentWord" -appID com.handsome.steganography -release

echo 打包完成！
pause