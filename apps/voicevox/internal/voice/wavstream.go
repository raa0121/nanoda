package voice

import "io"

// readAllAndClose はエンジンが返したWAVを全部メモリに読み出す。
// 合成結果はネイティブ側のメモリを指しているので、必ず読み切ってから
// Closeで解放する。
func readAllAndClose(r io.ReadCloser) ([]byte, error) {
	b, err := io.ReadAll(r)
	if cerr := r.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}
