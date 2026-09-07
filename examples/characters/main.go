package main

import (
	"fmt"

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

	// 読み込んでいるVVMのキャラクターを一覧する
	for _, c := range v.GetCharacters() {
		fmt.Printf("%s (%s)\n", c.Name, c.Id)
		for _, style := range c.Styles {
			fmt.Printf("\t%d: %s [%s]\n", style.Id, style.Name, style.Type)
		}
	}
}
