package voice

import (
	"errors"
	"fmt"

	nanoda2 "github.com/aethiopicuschan/nanoda/v2"
)

// Config はVOICEVOX COREを読み込むために必要なパス。
// CorePath が空の場合はデモエンジンが使われる。
type Config struct {
	// CorePath はコアライブラリ(voicevox_core.dll / libvoicevox_core.so /
	// libvoicevox_core.dylib)のパス。
	//
	// 0.16系ではONNX Runtimeをコアライブラリと同じディレクトリに
	// 置いておく必要がある。
	CorePath string
	// DictPath はOpenJTalkの辞書ディレクトリ。
	DictPath string
	// ModelPath は.vvmファイルが入ったディレクトリ。
	ModelPath string
}

// Open は設定に合ったエンジンを開く。
//
// バージョン判定は nanoda/v2 に任せている。v2 は voicevox_get_version() を
// 見て 0.16系 でなければ ErrUnsupportedVersion を返すので、それを合図に
// 0.15系用の nanoda (v1) へ切り替える。逆順に試すと、0.16系のコアに対して
// v1 が存在しないシンボルを引きに行ってpuregoがpanicする。
func Open(cfg Config) (Engine, error) {
	if cfg.CorePath == "" {
		return openDemo(), nil
	}
	e, err := open016(cfg)
	if err == nil {
		return e, nil
	}
	if !errors.Is(err, nanoda2.ErrUnsupportedVersion) {
		return nil, fmt.Errorf("VOICEVOX CORE 0.16系として読み込めませんでした: %w", err)
	}
	e, err015 := open015(cfg)
	if err015 != nil {
		return nil, fmt.Errorf("VOICEVOX CORE 0.16系として読み込めず、0.15系としても開けませんでした: %w", err015)
	}
	return e, nil
}
