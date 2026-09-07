package nanoda

import (
	"errors"
	"path/filepath"
	"runtime"
	"unsafe"
)

// ONNX Runtimeを読み込む手段が動的ライブラリに存在しない場合のエラー
var ErrOnnxruntimeUnavailable = errors.New("nanoda: this build of VOICEVOX CORE provides no way to load ONNX Runtime")

/*
ONNX Runtimeを読み込んで初期化する。
VOICEVOX CORE 0.16以降ではシンセサイザの構築にONNX Runtimeのインスタンスが必要となる。
ファイル名もディレクトリも指定されなかった場合は動的ライブラリが持つ既定のファイル名が使われる。
*/
func (v *Voicevox) initOnnxruntime(opt VoicevoxOption) (err error) {
	// 既に読み込まれている場合はそのインスタンスを使い回す
	if ptr := v.voicevoxOnnxruntimeGet(); ptr != 0 {
		v.onnxruntime = ptr
		return
	}
	filename := opt.onnxruntimeFilename
	// ディレクトリのみが指定された場合はファイル名を補完する
	if filename == "" && opt.onnxruntimeDir != "" {
		if v.voicevoxGetOnnxruntimeLibVersionedFilename == nil {
			err = ErrOnnxruntimeUnavailable
			return
		}
		filename = filepath.Join(opt.onnxruntimeDir, v.voicevoxGetOnnxruntimeLibVersionedFilename())
	}
	// ヌル終端した文字列がVOICEVOX COREから参照されている間、GCに回収されないようにする
	var nullTerminated []byte
	if filename != "" {
		nullTerminated = append([]byte(filename), 0)
	}
	defer runtime.KeepAlive(nullTerminated)

	var code ResultCode
	switch {
	// VOICEVOX_LOAD_ONNXRUNTIME付きでビルドされている場合
	case v.voicevoxOnnxruntimeLoadOnce != nil && v.voicevoxMakeDefaultLoadOnnxruntimeOptions != nil:
		// VoicevoxLoadOnnxruntimeOptionsはconst char *ひとつのみを持つ構造体
		options := v.voicevoxMakeDefaultLoadOnnxruntimeOptions()
		if nullTerminated != nil {
			options = uintptr(unsafe.Pointer(&nullTerminated[0]))
		}
		code = v.voicevoxOnnxruntimeLoadOnce(options, uintptr(unsafe.Pointer(&v.onnxruntime)))
	// VOICEVOX_LINK_ONNXRUNTIME付きでビルドされている場合
	case v.voicevoxOnnxruntimeInitOnce != nil:
		code = v.voicevoxOnnxruntimeInitOnce(uintptr(unsafe.Pointer(&v.onnxruntime)))
	default:
		err = ErrOnnxruntimeUnavailable
		return
	}
	if code != VOICEVOX_RESULT_OK {
		err = v.newError(code)
	}
	return
}

/*
ONNX Runtimeの動的ライブラリの、バージョン付きのファイル名を取得する。
WindowsとAndroidではGetOnnxruntimeLibUnversionedFilenameと同じ値を返す。
*/
func (v *Voicevox) GetOnnxruntimeLibVersionedFilename() (string, error) {
	if v.voicevoxGetOnnxruntimeLibVersionedFilename == nil {
		return "", ErrOnnxruntimeUnavailable
	}
	return v.voicevoxGetOnnxruntimeLibVersionedFilename(), nil
}

// ONNX Runtimeの動的ライブラリの、バージョン無しのファイル名を取得する
func (v *Voicevox) GetOnnxruntimeLibUnversionedFilename() (string, error) {
	if v.voicevoxGetOnnxruntimeLibUnversionedFilename == nil {
		return "", ErrOnnxruntimeUnavailable
	}
	return v.voicevoxGetOnnxruntimeLibUnversionedFilename(), nil
}
