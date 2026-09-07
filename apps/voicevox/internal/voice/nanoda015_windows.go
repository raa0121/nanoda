//go:build windows

package voice

import "errors"

// 0.15系のnanodaはpuregoのDlopenを直接呼んでおり、Windowsには実装が無い。
// (0.16系のnanoda/v2は内部にWindows用のLoadLibraryを持っているので動く。)
// そのため、Windowsでは0.15系へのフォールバックを行わない。
var errNo015OnWindows = errors.New("VOICEVOX CORE 0.15系はWindowsでは利用できません。0.16系のコアライブラリを指定してください")

func open015(cfg Config) (Engine, error) {
	return nil, errNo015OnWindows
}
