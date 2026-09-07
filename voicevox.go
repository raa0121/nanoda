package nanoda

import (
	"encoding/json"
	"unsafe"

	"github.com/aethiopicuschan/nanoda/internal/strings"
	"github.com/ebitengine/purego"
)

// 各種関数やポインタなどを保持する構造体
type Voicevox struct {
	core        uintptr
	onnxruntime uintptr
	openJtalkRc uintptr
	vvms        []voicevoxVoiceModel
	// 生(?)の関数群
	voicevoxAudioQueryCreateFromAccentPhrases      func(string, uintptr) ResultCode
	voicevoxErrorResultToMessage                   func(ResultCode) string
	voicevoxGetOnnxruntimeLibUnversionedFilename   func() string
	voicevoxGetOnnxruntimeLibVersionedFilename     func() string
	voicevoxGetVersion                             func() string
	voicevoxJsonFree                               func(uintptr)
	voicevoxMakeDefaultLoadOnnxruntimeOptions      func() uintptr
	voicevoxOnnxruntimeCreateSupportedDevicesJson  func(uintptr, uintptr) ResultCode
	voicevoxOnnxruntimeGet                         func() uintptr
	voicevoxOnnxruntimeInitOnce                    func(uintptr) ResultCode
	voicevoxOnnxruntimeLoadOnce                    func(uintptr, uintptr) ResultCode
	voicevoxOpenJtalkRcAnalyze                     func(uintptr, string, uintptr) ResultCode
	voicevoxOpenJtalkRcDelete                      func(uintptr)
	voicevoxOpenJtalkRcNew                         func(string, uintptr) ResultCode
	voicevoxOpenJtalkRcUseUserDict                 func(uintptr, uintptr) ResultCode
	voicevoxSynthesizerCreateAccentPhrases         func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerCreateAccentPhrasesFromKana func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerCreateAudioQuery            func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerCreateAudioQueryFromKana    func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerCreateMetasJson             func(uintptr) uintptr
	voicevoxSynthesizerDelete                      func(uintptr)
	voicevoxSynthesizerIsGpuMode                   func(uintptr) bool
	voicevoxSynthesizerIsLoadedVoiceModel          func(uintptr, uintptr) bool
	voicevoxSynthesizerLoadVoiceModel              func(uintptr, uintptr) ResultCode
	voicevoxSynthesizerNew                         func(uintptr, uintptr, uintptr, uintptr) ResultCode
	voicevoxSynthesizerReplaceMoraData             func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerReplaceMoraPitch            func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerReplacePhonemeLength        func(uintptr, string, StyleId, uintptr) ResultCode
	voicevoxSynthesizerSynthesis                   func(uintptr, string, StyleId, uintptr, uintptr, uintptr) ResultCode
	voicevoxSynthesizerTts                         func(uintptr, string, StyleId, uintptr, uintptr, uintptr) ResultCode
	voicevoxSynthesizerTtsFromKana                 func(uintptr, string, StyleId, uintptr, uintptr, uintptr) ResultCode
	voicevoxSynthesizerUnloadVoiceModel            func(uintptr, uintptr) ResultCode
	voicevoxUserDictAddWord                        func(uintptr, uintptr, uintptr) ResultCode
	voicevoxUserDictDelete                         func(uintptr)
	voicevoxUserDictImport                         func(uintptr, uintptr) ResultCode
	voicevoxUserDictLoad                           func(uintptr, string) ResultCode
	voicevoxUserDictNew                            func() uintptr
	voicevoxUserDictRemoveWord                     func(uintptr, uintptr) ResultCode
	voicevoxUserDictSave                           func(uintptr, string) ResultCode
	voicevoxUserDictToJson                         func(uintptr, uintptr) ResultCode
	voicevoxUserDictUpdateWord                     func(uintptr, uintptr, uintptr) ResultCode
	voicevoxVoiceModelFileCreateMetasJson          func(uintptr) uintptr
	voicevoxVoiceModelFileDelete                   func(uintptr)
	voicevoxVoiceModelFileId                       func(uintptr, uintptr)
	voicevoxVoiceModelFileOpen                     func(string, uintptr) ResultCode
	voicevoxWavFree                                func(uintptr)
}

// Voicevoxの生成時に指定するオプション
type VoicevoxOption struct {
	onnxruntimeFilename string
	onnxruntimeDir      string
}

/*
ONNX Runtimeのファイル名(モジュール名)もしくはファイルパスを指定する
指定しない場合はVOICEVOX COREが持つ既定のファイル名が使われる
*/
func WithOnnxruntime(filename string) func(*VoicevoxOption) {
	return func(o *VoicevoxOption) {
		o.onnxruntimeFilename = filename
	}
}

/*
ONNX Runtimeの動的ライブラリが置かれているディレクトリを指定する
ファイル名はVOICEVOX COREが持つ既定のものが使われるため、
ダウンローダーが展開したvoicevox_core/onnxruntime/libをそのまま渡せる
WithOnnxruntimeが指定されている場合はそちらが優先される
*/
func WithOnnxruntimeDir(dir string) func(*VoicevoxOption) {
	return func(o *VoicevoxOption) {
		o.onnxruntimeDir = dir
	}
}

// 必要なパスを引数に取り、Voicevoxのインスタンスを生成する
func NewVoicevox(corePath string, openJtalkPath string, modelPath string, options ...func(*VoicevoxOption)) (v *Voicevox, err error) {
	opt := VoicevoxOption{}
	for _, o := range options {
		o(&opt)
	}
	c, err := purego.Dlopen(corePath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return
	}
	v = &Voicevox{
		core: c,
	}
	// 関数の紐付けを行う
	v.register()
	// ONNX Runtimeを読み込む
	if err = v.initOnnxruntime(opt); err != nil {
		return
	}
	// OpenJtalkRcを構築する
	code := v.voicevoxOpenJtalkRcNew(openJtalkPath, uintptr(unsafe.Pointer(&v.openJtalkRc)))
	if code != VOICEVOX_RESULT_OK {
		err = v.newError(code)
		return
	}
	// VVMファイルを読み込む
	v.vvms, err = v.loadVVMs(modelPath)
	// エラーがあった場合、もろもろを破棄する
	if err != nil {
		v.voicevoxOpenJtalkRcDelete(v.openJtalkRc)
		v.deleteVVMs()
	}
	return
}

// 利用可能なデバイスの情報を取得する
func (v *Voicevox) SupportedDevices() (sd SupportedDevices, err error) {
	var ptr *byte
	code := v.voicevoxOnnxruntimeCreateSupportedDevicesJson(v.onnxruntime, uintptr(unsafe.Pointer(&ptr)))
	if code != VOICEVOX_RESULT_OK {
		err = v.newError(code)
		return
	}
	defer v.voicevoxJsonFree(uintptr(unsafe.Pointer(ptr)))
	j := strings.GoString(ptr)
	err = json.Unmarshal([]byte(j), &sd)
	return
}

// 結果コードからメッセージを取得する
func (v *Voicevox) GetMessageFromResult(code ResultCode) string {
	return v.voicevoxErrorResultToMessage(code)
}

// voicevoxのバージョンを取得する
func (v *Voicevox) GetVersion() string {
	return v.voicevoxGetVersion()
}

// 読み込んでいるVVMのキャラクターの一覧を取得する
func (v *Voicevox) GetCharacters() []Character {
	return mergeCharacters(v.getCharacterMetas())
}

// メタ情報の一覧を取得する
//
// Deprecated: VOICEVOX CORE 0.16で「話者」は「キャラクター」に改称された。GetCharactersを使うこと。
func (v *Voicevox) GetMetas() []Meta {
	return toMetas(v.GetCharacters())
}

/*
スタイルの一覧を取得する
ここで取得されるスタイルのNameは「ずんだもん(あまあま)」のようなフォーマットになる
*/
func (v *Voicevox) GetStyles() []Style {
	return toStyles(v.GetCharacters())
}
