package nanoda

import (
	"encoding/json"
	"unsafe"

	"github.com/aethiopicuschan/nanoda/internal/strings"
)

/*
音声合成用のクエリ
VOICEVOX CORE 0.16では各スケールのJSONフィールド名がキャメルケースとなっている
*/
type AudioQuery struct {
	AccentPhrases      []AccentPhrase `json:"accent_phrases"`
	SpeedScale         float64        `json:"speedScale"`
	PitchScale         float64        `json:"pitchScale"`
	IntonationScale    float64        `json:"intonationScale"`
	VolumeScale        float64        `json:"volumeScale"`
	PrePhonemeLength   float64        `json:"prePhonemeLength"`
	PostPhonemeLength  float64        `json:"postPhonemeLength"`
	OutputSamplingRate int            `json:"outputSamplingRate"`
	OutputStereo       bool           `json:"outputStereo"`
	// AquesTalk風記法 存在しない場合はnil
	Kana *string `json:"kana"`
}

// 音声合成用のクエリを作成する
func (s *Synthesizer) createAudioQuery(text string, styleID StyleId, enableKana bool) (a AudioQuery, err error) {
	var ptr *byte
	var code ResultCode
	if enableKana {
		code = s.v.voicevoxSynthesizerCreateAudioQueryFromKana(s.synthesizer, text, styleID, uintptr(unsafe.Pointer(&ptr)))
	} else {
		code = s.v.voicevoxSynthesizerCreateAudioQuery(s.synthesizer, text, styleID, uintptr(unsafe.Pointer(&ptr)))
	}
	if code != VOICEVOX_RESULT_OK {
		err = s.v.newError(code)
		return
	}
	defer s.v.voicevoxJsonFree(uintptr(unsafe.Pointer(ptr)))
	j := strings.GoString(ptr)
	err = json.Unmarshal([]byte(j), &a)
	return
}

// 音声合成用のクエリを生成する
func (s *Synthesizer) CreateAudioQuery(text string, styleID StyleId) (a AudioQuery, err error) {
	return s.createAudioQuery(text, styleID, false)
}

// 音声合成用のクエリを生成する(AquesTalk風記法)
func (s *Synthesizer) CreateAudioQueryFromKana(text string, styleID StyleId) (a AudioQuery, err error) {
	return s.createAudioQuery(text, styleID, true)
}

/*
アクセント句の配列から音声合成用のクエリを生成する
アクセントを編集したアクセント句からクエリを作る際に使う
*/
func (v *Voicevox) CreateAudioQueryFromAccentPhrases(ap []AccentPhrase) (a AudioQuery, err error) {
	jB, err := json.Marshal(ap)
	if err != nil {
		return
	}
	var ptr *byte
	code := v.voicevoxAudioQueryCreateFromAccentPhrases(string(jB), uintptr(unsafe.Pointer(&ptr)))
	if code != VOICEVOX_RESULT_OK {
		err = v.newError(code)
		return
	}
	defer v.voicevoxJsonFree(uintptr(unsafe.Pointer(ptr)))
	j := strings.GoString(ptr)
	err = json.Unmarshal([]byte(j), &a)
	return
}

/*
アクセント句の配列から音声合成用のクエリを生成する
アクセントを編集したアクセント句からクエリを作る際に使う
*/
func (s *Synthesizer) CreateAudioQueryFromAccentPhrases(ap []AccentPhrase) (a AudioQuery, err error) {
	return s.v.CreateAudioQueryFromAccentPhrases(ap)
}
