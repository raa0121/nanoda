package nanoda

import (
	"encoding/json"
	"unsafe"

	"github.com/aethiopicuschan/nanoda/internal/strings"
)

// モーラ（子音＋母音）
type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

// アクセント句
type AccentPhrase struct {
	Moras []Mora `json:"moras"`
	// アクセント位置 1始まりで、音が下がる直前のモーラを指す
	Accent          int   `json:"accent"`
	PauseMora       *Mora `json:"pause_mora"`
	IsInterrogative bool  `json:"is_interrogative"`
}

/*
日本語のテキストを解析してアクセント句の配列を生成する
音高・音素長は生成されないため、音声合成に用いる際はReplaceを呼んで生成しなおす必要がある
シンセサイザを必要としないため、音声モデルを読み込む前でも利用できる
*/
func (v *Voicevox) Analyze(text string) (a []AccentPhrase, err error) {
	var ptr *byte
	code := v.voicevoxOpenJtalkRcAnalyze(v.openJtalkRc, text, uintptr(unsafe.Pointer(&ptr)))
	if code != VOICEVOX_RESULT_OK {
		err = v.newError(code)
		return
	}
	defer v.voicevoxJsonFree(uintptr(unsafe.Pointer(ptr)))
	j := strings.GoString(ptr)
	err = json.Unmarshal([]byte(j), &a)
	return
}

// アクセント句の配列を生成する
func (s *Synthesizer) createAccentPhrases(text string, styleID StyleId, enableKana bool) (a []AccentPhrase, err error) {
	var ptr *byte
	var code ResultCode
	if enableKana {
		code = s.v.voicevoxSynthesizerCreateAccentPhrasesFromKana(s.synthesizer, text, styleID, uintptr(unsafe.Pointer(&ptr)))
	} else {
		code = s.v.voicevoxSynthesizerCreateAccentPhrases(s.synthesizer, text, styleID, uintptr(unsafe.Pointer(&ptr)))
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

// アクセント句の配列を生成する
func (s *Synthesizer) CreateAccentPhrases(text string, styleID StyleId) (a []AccentPhrase, err error) {
	return s.createAccentPhrases(text, styleID, false)
}

// アクセント句の配列を生成する(AquesTalk風記法)
func (s *Synthesizer) CreateAccentPhrasesFromKana(text string, styleID StyleId) (a []AccentPhrase, err error) {
	return s.createAccentPhrases(text, styleID, true)
}
