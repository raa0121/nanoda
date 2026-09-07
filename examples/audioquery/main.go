package main

import (
	"io"
	"os"

	"github.com/aethiopicuschan/nanoda"
)

func main() {
	v, _ := nanoda.NewVoicevox(
		"voicevox_core/c_api/lib/libvoicevox_core.dylib",
		"voicevox_core/dict/open_jtalk_dic_utf_8-1.11",
		"voicevox_core/models/vvms",
		nanoda.WithOnnxruntimeDir("voicevox_core/onnxruntime/lib"),
	)
	s, _ := v.NewSynthesizer()
	s.LoadModelsFromStyleId(3)

	aq, _ := s.CreateAudioQuery("2倍速ずんだもんなのだ！", 3)
	aq.SpeedScale = 2.0

	wav, _ := s.Synthesis(aq, 3)
	defer wav.Close()
	f, _ := os.Create("output.wav")
	defer f.Close()
	io.Copy(f, wav)
}
