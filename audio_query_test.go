package nanoda

import (
	"encoding/json"
	"testing"
)

// VOICEVOX CORE 0.16のAudioQueryのJSONを扱えることを確認する
func TestAudioQueryUnmarshal(t *testing.T) {
	raw := `{
		"accent_phrases": [
			{
				"moras": [
					{"text": "ア", "consonant": null, "consonant_length": null, "vowel": "a", "vowel_length": 0.1, "pitch": 5.5}
				],
				"accent": 1,
				"pause_mora": null,
				"is_interrogative": false
			}
		],
		"speedScale": 1.0,
		"pitchScale": 0.0,
		"intonationScale": 1.0,
		"volumeScale": 1.0,
		"prePhonemeLength": 0.1,
		"postPhonemeLength": 0.1,
		"outputSamplingRate": 24000,
		"outputStereo": false,
		"kana": null
	}`

	var aq AudioQuery
	if err := json.Unmarshal([]byte(raw), &aq); err != nil {
		t.Fatal(err)
	}
	if aq.SpeedScale != 1.0 || aq.OutputSamplingRate != 24000 {
		t.Errorf("unexpected audio query: %+v", aq)
	}
	if aq.Kana != nil {
		t.Errorf("expected nil kana, got %q", *aq.Kana)
	}
	if len(aq.AccentPhrases) != 1 || aq.AccentPhrases[0].Accent != 1 {
		t.Fatalf("unexpected accent phrases: %+v", aq.AccentPhrases)
	}
	mora := aq.AccentPhrases[0].Moras[0]
	if mora.Consonant != nil || mora.ConsonantLength != nil {
		t.Errorf("unexpected consonant: %+v", mora)
	}
	if mora.Vowel != "a" || mora.Pitch != 5.5 {
		t.Errorf("unexpected mora: %+v", mora)
	}
}

// スケールがキャメルケースで出力されることを確認する
func TestAudioQueryMarshal(t *testing.T) {
	kana := "ア'"
	aq := AudioQuery{
		AccentPhrases:      []AccentPhrase{},
		SpeedScale:         2.0,
		OutputSamplingRate: 24000,
		Kana:               &kana,
	}
	b, err := json.Marshal(aq)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]any{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"accent_phrases", "speedScale", "pitchScale", "intonationScale", "volumeScale", "prePhonemeLength", "postPhonemeLength", "outputSamplingRate", "outputStereo", "kana"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing key: %s", key)
		}
	}
	if m["speedScale"] != 2.0 {
		t.Errorf("unexpected speedScale: %v", m["speedScale"])
	}
}
