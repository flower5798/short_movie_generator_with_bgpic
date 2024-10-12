package main

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sync"
)

func timeToFloat64Sec(timeStr string) (float64, error) {
	// Split the time string by ":"
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	// Convert hours, minutes, and seconds to float64
	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	// Convert the time to total seconds
	totalSeconds := hours*3600 + minutes*60 + seconds
	return totalSeconds, nil
}

func generateImageFrame(i int, outputPicNumber int, bgImage image.Image, bgWidth int, bgHeight int, wg *sync.WaitGroup) {
	defer wg.Done() // このgoroutineが終了したらWaitGroupのカウンタをデクリメント

	// 出力用のキャンバスを作成
	output := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))

	// 背景画像をキャンバスに描画
	draw.Draw(output, output.Bounds(), bgImage, image.Point{}, draw.Src)

	// 連番のファイル名を作成
	filename := fmt.Sprintf("decomposed_ffmpeg_frames/frame_%d.png", i)

	// 画像を読み込む
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("画像ファイルを開けません:", err)
		return
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		fmt.Println("画像のデコードに失敗しました:", err)
		return
	}

	// 読み込んだ画像のサイズ
	imgBounds := img.Bounds()
	imgWidth := imgBounds.Dx()
	imgHeight := imgBounds.Dy()

	// 拡大縮小の比率を計算（横幅を背景画像に合わせる）
	picZoomCoeff := 1.3
	scale := float64(bgWidth) / float64(imgWidth)
	newWidth := uint(float64(bgWidth) * picZoomCoeff)
	newHeight := uint(float64(imgHeight) * picZoomCoeff * scale)

	// 画像のリサイズを行う
	resizedImage := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// 背景にリサイズした画像を貼り付け
	offset := image.Pt((bgWidth-int(newWidth))/2, (bgHeight-int(newHeight))/2) // 中央に配置
	draw.Draw(output, resizedImage.Bounds().Add(offset), resizedImage, image.Point{}, draw.Over)

	// 出力ファイルを保存
	outFilename := fmt.Sprintf("edited_movie_frames/frame_%04d.png", outputPicNumber)
	outFile, err := os.Create(outFilename)
	if err != nil {
		fmt.Println("出力ファイルを作成できません:", err)
		return
	}
	defer outFile.Close()

	// PNG形式で保存
	png.Encode(outFile, output)
}

func main() {
	// 引数読み込み
	args := os.Args
	if len(args) != 4 {
		panic(fmt.Sprintf("number of args is not correct. args length: %d", len(args)))
	}
	backgroundImageFileName := args[1]
	startTime := args[2]
	endTime := args[3]

	startSecFloat64, _ := timeToFloat64Sec(startTime)
	endSecFloat64, _ := timeToFloat64Sec(endTime)

	fmt.Printf("background image file name: %s\n", backgroundImageFileName)
	fmt.Printf("duration: %f to %f\n", startSecFloat64, endSecFloat64)

	fps := 30 // TODO 任意引数化

	// 背景画像を読み込む
	bgFile, err := os.Open(backgroundImageFileName)
	if err != nil {
		fmt.Println("背景画像を開けません:", err)
		return
	}
	defer bgFile.Close()

	bgImage, err := png.Decode(bgFile)
	if err != nil {
		fmt.Println("背景画像のデコードに失敗しました。背景画像にはpngファイルを指定してください。\n", err)
		return
	}

	// 背景画像のサイズ
	bgBounds := bgImage.Bounds()
	bgWidth := bgBounds.Dx()
	bgHeight := bgBounds.Dy()

	// 連番のPNGファイルを読み込んで、背景に貼り付ける
	startFrameNumber := max(int(float64(fps) * startSecFloat64), 1) // ffmpegで吐き出される画像ファイルの連番の開始は1

	// frameのファイル数をカウント
	pattern := "decomposed_ffmpeg_frames/frame_*.png"
	files, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}
	frameFilesCount := len(files)
	fmt.Printf("frame files count: %d\n", frameFilesCount)

	endFrameNumber := min(int(float64(fps) * endSecFloat64), frameFilesCount)

	// goroutineを使った変換処理を実行
	var wg sync.WaitGroup
	fmt.Println("converting...")
	for i := startFrameNumber; i <= endFrameNumber; i++ {
		wg.Add(1) // WaitGroupのカウンタをインクリメント
		outputPicNumber := i - startFrameNumber
		go generateImageFrame(i, outputPicNumber, bgImage, bgWidth, bgHeight, &wg) // goroutineでフレーム生成
	}

	// 全てのgoroutineが完了するまで待機
	wg.Wait()

	fmt.Println("処理が完了しました。")
}
