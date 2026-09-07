// nanoda-voicevox は guigui と nanoda を使ってVOICEVOXのエディタを
// 再現したデスクトップアプリケーション。
//
//	go run . -core /path/to/voicevox_core.dll \
//	         -dict /path/to/open_jtalk_dic_utf_8-1.11 \
//	         -models /path/to/models
//
// -core を省略するとデモエンジンで起動し、VOICEVOX CORE 無しでも
// 画面と操作を確認できる。
package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/guigui-gui/guigui"
	// 日本語を表示するために必要。
	_ "github.com/guigui-gui/guigui/basicwidget/cjkfont"

	"github.com/raa0121/nanoda/apps/voicevox/internal/player"
	"github.com/raa0121/nanoda/apps/voicevox/internal/project"
	"github.com/raa0121/nanoda/apps/voicevox/internal/ui"
	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

func main() {
	corePath := flag.String("core", os.Getenv("NANODA_CORE"),
		"VOICEVOX COREのライブラリ(voicevox_core.dll / libvoicevox_core.so / .dylib)のパス。省略するとデモエンジンで起動する")
	dictPath := flag.String("dict", os.Getenv("NANODA_DICT"),
		"OpenJTalkの辞書ディレクトリ")
	modelPath := flag.String("models", os.Getenv("NANODA_MODELS"),
		"音声モデル(.vvm)が入ったディレクトリ")
	outDir := flag.String("out", "voicevox-out",
		"WAVとプロジェクトの書き出し先ディレクトリ")
	openPath := flag.String("open", "",
		"起動時に開くプロジェクトファイル")
	flag.Parse()

	if err := run(*corePath, *dictPath, *modelPath, *outDir, *openPath); err != nil {
		log.Fatal(err)
	}
}

func run(corePath, dictPath, modelPath, outDir, openPath string) error {
	engine, err := voice.Open(voice.Config{
		CorePath:  corePath,
		DictPath:  dictPath,
		ModelPath: modelPath,
	})
	if err != nil {
		return err
	}

	proj := project.New(voice.FirstStyleID(engine.Speakers()))
	if openPath != "" {
		proj, err = project.Load(openPath)
		if err != nil {
			return fmt.Errorf("%s を開けませんでした: %w", openPath, err)
		}
	}

	app := ui.NewApp(engine, player.New(), proj, outDir)
	defer app.Close()

	return guigui.Run(ui.NewRoot(app), &guigui.RunOptions{
		Title:         "nanoda VOICEVOX",
		WindowSize:    image.Pt(1280, 800),
		WindowMinSize: image.Pt(1000, 640),
	})
}
