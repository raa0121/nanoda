# nanoda

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/aethiopicuschan/nanoda.svg)](https://pkg.go.dev/github.com/aethiopicuschan/nanoda)
[![CI](https://github.com/aethiopicuschan/nanoda/actions/workflows/ci.yml/badge.svg)](https://github.com/aethiopicuschan/nanoda/actions/workflows/ci.yml)

nanodaは[VOICEVOX CORE](https://github.com/VOICEVOX/voicevox_core)の動的ライブラリをGolangから叩くためのライブラリです。`cgo`ではなく[ebitengine/purego](https://github.com/ebitengine/purego/)を利用しているため、簡単に使用することが可能です。

## VOICEVOXについて

サポートするVOICEVOX COREのバージョンは `0.16` としており、開発は `0.16.0` を元にしています。

nanoda自体は[MITライセンス](/LICENSE)ですが、利用に際してはVOICEVOXやOpenJTalkの利用規約に則る必要があることに注意してください。

## 使い方

```sh
go get github.com/aethiopicuschan/nanoda@latest
```

もっとも簡単な例は以下のようになります。

```go
v, _ := nanoda.NewVoicevox(
	"voicevox_core/c_api/lib/libvoicevox_core.dylib",
	"voicevox_core/dict/open_jtalk_dic_utf_8-1.11",
	"voicevox_core/models/vvms",
	nanoda.WithOnnxruntimeDir("voicevox_core/onnxruntime/lib"),
)
s, _ := v.NewSynthesizer()
s.LoadAllModels()
wav, _ := s.Tts("ずんだもんなのだ！", 3)
defer wav.Close()
f, _ := os.Create("output.wav")
defer f.Close()
io.Copy(f, wav)
```

0.16では音声合成にONNX Runtimeが必要となったため、`NewVoicevox` の中でその読み込みも行っています。ダウンローダーが展開したディレクトリをそのまま `WithOnnxruntimeDir` に渡すか、ファイル名やパスを `WithOnnxruntime` に渡してください。どちらも指定しない場合はVOICEVOX COREが持つ既定のファイル名で読み込まれます。

### キャラクター一覧

読み込んでいるVVMのキャラクターとそのスタイルは `GetCharacters` で取得できます。並び順はVOICEVOX CORE本体と同じく `order` の昇順です。

```go
for _, c := range v.GetCharacters() {
	fmt.Printf("%s (%s)\n", c.Name, c.Id)
	for _, style := range c.Styles {
		fmt.Printf("\t%d: %s [%s]\n", style.Id, style.Name, style.Type)
	}
}
```

シンセサイザに読み込み済みのものだけが欲しい場合は `Synthesizer.GetCharacters` を使ってください。「ずんだもん(あまあま)」のような表示名が欲しい場合は `GetStyles` が便利です。

### AudioQueryとアクセント編集

`CreateAudioQuery` で得た `AudioQuery` を編集して `Synthesis` に渡す方法のほかに、アクセント句を編集してから `AudioQuery` を組み立てることもできます。

```go
// テキストを解析してアクセント句を得る(音高・音素長は未生成)
ap, _ := v.Analyze("ずんだもんなのだ！")

// アクセントを編集する
ap[0].Accent = 1
ap[len(ap)-1].IsInterrogative = true

// 編集後のアクセント句に対して音高・音素長を生成しなおす
ap, _ = s.Replace(ap, 3)

// アクセント句からAudioQueryを組み立てる
aq, _ := s.CreateAudioQueryFromAccentPhrases(ap)
wav, _ := s.Synthesis(aq, 3)
```

`Analyze` はシンセサイザを必要としないため、音声モデルを読み込む前でも利用できます。音高のみ・音素長のみを生成しなおしたい場合は `ReplaceOnlyMoraPitch` ・ `ReplaceOnlyPhonemeLength` を使ってください。

その他 `examples` ディレクトリにサンプルコードを置いていますので、ご活用ください。

## 0.15からの変更点

- 「話者」が「キャラクター」に改称されました。 `GetMetas` や `LoadModelsFromSpeakerId` などは残していますが、 `GetCharacters` や `LoadModelsFromCharacterId` の利用を推奨します。
- `Style` に `Type` と `Order` が増えました。
- `AudioQuery` の各スケールのJSONフィールド名がキャメルケースになりました。Goの構造体としてのフィールド名は変わっていません。
- `AudioQuery.Kana` は省略可能となったため `*string` になりました。
- ONNX Runtimeの読み込みが必要になりました。

## 動作環境

```
GOARCH='arm64'
GOOS='darwin'
```

でのみ確認しています。

## 開発方針

以下の理由からなるべくnanoda側で処理を受け持ったり抽象化したりして機能を提供することを目指しています。

- 使いやすさの向上
- メモリまわりの安全性要件の確保
- VOICEVOXとアプリケーション間の密結合を避け、APIの変更等に強くする

## テスト

動的ライブラリを必要としない範囲（メタ情報のマージやJSONの変換）についてのみテストがあります。

```sh
go test ./...
```

## 対応状況

以下は内部的に利用している関数のリストであり、必ずしも一致する形で公開されているわけではありません。

- [x] voicevox_audio_query_create_from_accent_phrases
- [x] voicevox_error_result_to_message
- [x] voicevox_get_onnxruntime_lib_unversioned_filename
- [x] voicevox_get_onnxruntime_lib_versioned_filename
- [x] voicevox_get_version
- [x] voicevox_json_free
- [x] voicevox_make_default_load_onnxruntime_options
- [ ] voicevox_make_default_initialize_options
- [ ] voicevox_make_default_synthesis_options
- [ ] voicevox_make_default_tts_options
- [x] voicevox_onnxruntime_create_supported_devices_json
- [x] voicevox_onnxruntime_get
- [x] voicevox_onnxruntime_init_once
- [x] voicevox_onnxruntime_load_once
- [x] voicevox_open_jtalk_rc_analyze
- [x] voicevox_open_jtalk_rc_delete
- [x] voicevox_open_jtalk_rc_new
- [x] voicevox_open_jtalk_rc_use_user_dict
- [x] voicevox_synthesizer_create_accent_phrases
- [x] voicevox_synthesizer_create_accent_phrases_from_kana
- [x] voicevox_synthesizer_create_audio_query
- [x] voicevox_synthesizer_create_audio_query_from_kana
- [x] voicevox_synthesizer_create_metas_json
- [x] voicevox_synthesizer_delete
- [ ] voicevox_synthesizer_get_onnxruntime
- [x] voicevox_synthesizer_is_gpu_mode
- [x] voicevox_synthesizer_is_loaded_voice_model
- [x] voicevox_synthesizer_load_voice_model
- [x] voicevox_synthesizer_new
- [x] voicevox_synthesizer_replace_mora_data
- [x] voicevox_synthesizer_replace_mora_pitch
- [x] voicevox_synthesizer_replace_phoneme_length
- [x] voicevox_synthesizer_synthesis
- [x] voicevox_synthesizer_tts
- [x] voicevox_synthesizer_tts_from_kana
- [x] voicevox_synthesizer_unload_voice_model
- [x] voicevox_user_dict_add_word
- [x] voicevox_user_dict_delete
- [x] voicevox_user_dict_import
- [x] voicevox_user_dict_load
- [x] voicevox_user_dict_new
- [x] voicevox_user_dict_remove_word
- [x] voicevox_user_dict_save
- [x] voicevox_user_dict_to_json
- [x] voicevox_user_dict_update_word
- [ ] voicevox_user_dict_word_make
- [x] voicevox_voice_model_file_create_metas_json
- [x] voicevox_voice_model_file_delete
- [x] voicevox_voice_model_file_id
- [x] voicevox_voice_model_file_open
- [x] voicevox_wav_free
