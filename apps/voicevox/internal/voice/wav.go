package voice

import (
	"encoding/binary"
	"math"
)

// encodeWAV は16bitモノラルPCMのRIFF WAVEを組み立てる。
// VOICEVOX COREの出力と同じ形式なので、再生・書き出しの経路を
// エンジンの種類によらず一本にできる。
func encodeWAV(samples []float64, sampleRate int) []byte {
	const bitsPerSample = 16
	const channels = 1
	dataSize := len(samples) * 2
	buf := make([]byte, 0, 44+dataSize)

	buf = append(buf, "RIFF"...)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(36+dataSize))
	buf = append(buf, "WAVE"...)

	buf = append(buf, "fmt "...)
	buf = binary.LittleEndian.AppendUint32(buf, 16)
	buf = binary.LittleEndian.AppendUint16(buf, 1) // PCM
	buf = binary.LittleEndian.AppendUint16(buf, channels)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(sampleRate))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(sampleRate*channels*bitsPerSample/8))
	buf = binary.LittleEndian.AppendUint16(buf, channels*bitsPerSample/8)
	buf = binary.LittleEndian.AppendUint16(buf, bitsPerSample)

	buf = append(buf, "data"...)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(dataSize))
	for _, s := range samples {
		buf = binary.LittleEndian.AppendUint16(buf, uint16(int16(math.Round(clamp(s, -1, 1)*32767))))
	}
	return buf
}

func clamp(v, lo, hi float64) float64 {
	return math.Min(math.Max(v, lo), hi)
}
