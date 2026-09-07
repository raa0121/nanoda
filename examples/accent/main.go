package main

import (
	"io"
	"os"

	"github.com/aethiopicuschan/nanoda"
)

func main() {
	v, err := nanoda.NewVoicevox(
		"voicevox_core/c_api/lib/libvoicevox_core.dylib",
		"voicevox_core/dict/open_jtalk_dic_utf_8-1.11",
		"voicevox_core/models/vvms",
		nanoda.WithOnnxruntimeDir("voicevox_core/onnxruntime/lib"),
	)
	if err != nil {
		panic(err)
	}
	s, err := v.NewSynthesizer()
	if err != nil {
		panic(err)
	}
	defer s.Close()
	if err := s.LoadModelsFromStyleId(3); err != nil {
		panic(err)
	}

	// テキストを解析してアクセント句を得る
	// この時点では音高・音素長は生成されていない
	ap, err := v.Analyze("ずんだもんなのだ！")
	if err != nil {
		panic(err)
	}

	// アクセントを編集する
	// 疑問文として読み上げさせつつ、最初のアクセント句のアクセント位置を先頭に移す
	ap[0].Accent = 1
	ap[len(ap)-1].IsInterrogative = true

	// 編集後のアクセント句に対して音高・音素長を生成しなおす
	ap, err = s.Replace(ap, 3)
	if err != nil {
		panic(err)
	}

	// アクセント句からAudioQueryを組み立てて音声合成する
	aq, err := s.CreateAudioQueryFromAccentPhrases(ap)
	if err != nil {
		panic(err)
	}
	wav, err := s.Synthesis(aq, 3)
	if err != nil {
		panic(err)
	}
	defer wav.Close()

	f, err := os.Create("output.wav")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, wav); err != nil {
		panic(err)
	}
}
