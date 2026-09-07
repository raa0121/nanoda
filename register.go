package nanoda

import (
	"github.com/ebitengine/purego"
)

// 関数の紐付けを行う
func (v *Voicevox) register() {
	// ONNX Runtime
	// VOICEVOX_LOAD_ONNXRUNTIME/VOICEVOX_LINK_ONNXRUNTIMEのどちらでビルドされているかによって
	// 存在するシンボルが異なるため、存在するものだけを紐付ける
	v.registerOptional(&v.voicevoxGetOnnxruntimeLibVersionedFilename, "voicevox_get_onnxruntime_lib_versioned_filename")
	v.registerOptional(&v.voicevoxGetOnnxruntimeLibUnversionedFilename, "voicevox_get_onnxruntime_lib_unversioned_filename")
	v.registerOptional(&v.voicevoxMakeDefaultLoadOnnxruntimeOptions, "voicevox_make_default_load_onnxruntime_options")
	v.registerOptional(&v.voicevoxOnnxruntimeLoadOnce, "voicevox_onnxruntime_load_once")
	v.registerOptional(&v.voicevoxOnnxruntimeInitOnce, "voicevox_onnxruntime_init_once")
	purego.RegisterLibFunc(&v.voicevoxOnnxruntimeGet, v.core, "voicevox_onnxruntime_get")
	purego.RegisterLibFunc(&v.voicevoxOnnxruntimeCreateSupportedDevicesJson, v.core, "voicevox_onnxruntime_create_supported_devices_json")
	// Open JTalk
	purego.RegisterLibFunc(&v.voicevoxOpenJtalkRcNew, v.core, "voicevox_open_jtalk_rc_new")
	purego.RegisterLibFunc(&v.voicevoxOpenJtalkRcUseUserDict, v.core, "voicevox_open_jtalk_rc_use_user_dict")
	purego.RegisterLibFunc(&v.voicevoxOpenJtalkRcAnalyze, v.core, "voicevox_open_jtalk_rc_analyze")
	purego.RegisterLibFunc(&v.voicevoxOpenJtalkRcDelete, v.core, "voicevox_open_jtalk_rc_delete")
	// 音声モデルファイル
	purego.RegisterLibFunc(&v.voicevoxVoiceModelFileOpen, v.core, "voicevox_voice_model_file_open")
	purego.RegisterLibFunc(&v.voicevoxVoiceModelFileId, v.core, "voicevox_voice_model_file_id")
	purego.RegisterLibFunc(&v.voicevoxVoiceModelFileCreateMetasJson, v.core, "voicevox_voice_model_file_create_metas_json")
	purego.RegisterLibFunc(&v.voicevoxVoiceModelFileDelete, v.core, "voicevox_voice_model_file_delete")
	// シンセサイザ
	purego.RegisterLibFunc(&v.voicevoxSynthesizerNew, v.core, "voicevox_synthesizer_new")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerDelete, v.core, "voicevox_synthesizer_delete")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerLoadVoiceModel, v.core, "voicevox_synthesizer_load_voice_model")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerUnloadVoiceModel, v.core, "voicevox_synthesizer_unload_voice_model")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerIsGpuMode, v.core, "voicevox_synthesizer_is_gpu_mode")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerIsLoadedVoiceModel, v.core, "voicevox_synthesizer_is_loaded_voice_model")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerCreateMetasJson, v.core, "voicevox_synthesizer_create_metas_json")
	// AudioQuery
	purego.RegisterLibFunc(&v.voicevoxAudioQueryCreateFromAccentPhrases, v.core, "voicevox_audio_query_create_from_accent_phrases")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerCreateAudioQuery, v.core, "voicevox_synthesizer_create_audio_query")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerCreateAudioQueryFromKana, v.core, "voicevox_synthesizer_create_audio_query_from_kana")
	// アクセント句
	purego.RegisterLibFunc(&v.voicevoxSynthesizerCreateAccentPhrases, v.core, "voicevox_synthesizer_create_accent_phrases")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerCreateAccentPhrasesFromKana, v.core, "voicevox_synthesizer_create_accent_phrases_from_kana")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerReplaceMoraData, v.core, "voicevox_synthesizer_replace_mora_data")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerReplaceMoraPitch, v.core, "voicevox_synthesizer_replace_mora_pitch")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerReplacePhonemeLength, v.core, "voicevox_synthesizer_replace_phoneme_length")
	// 音声合成
	purego.RegisterLibFunc(&v.voicevoxSynthesizerSynthesis, v.core, "voicevox_synthesizer_synthesis")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerTts, v.core, "voicevox_synthesizer_tts")
	purego.RegisterLibFunc(&v.voicevoxSynthesizerTtsFromKana, v.core, "voicevox_synthesizer_tts_from_kana")
	// ユーザー辞書
	purego.RegisterLibFunc(&v.voicevoxUserDictAddWord, v.core, "voicevox_user_dict_add_word")
	purego.RegisterLibFunc(&v.voicevoxUserDictDelete, v.core, "voicevox_user_dict_delete")
	purego.RegisterLibFunc(&v.voicevoxUserDictImport, v.core, "voicevox_user_dict_import")
	purego.RegisterLibFunc(&v.voicevoxUserDictLoad, v.core, "voicevox_user_dict_load")
	purego.RegisterLibFunc(&v.voicevoxUserDictNew, v.core, "voicevox_user_dict_new")
	purego.RegisterLibFunc(&v.voicevoxUserDictRemoveWord, v.core, "voicevox_user_dict_remove_word")
	purego.RegisterLibFunc(&v.voicevoxUserDictSave, v.core, "voicevox_user_dict_save")
	purego.RegisterLibFunc(&v.voicevoxUserDictToJson, v.core, "voicevox_user_dict_to_json")
	purego.RegisterLibFunc(&v.voicevoxUserDictUpdateWord, v.core, "voicevox_user_dict_update_word")
	// その他
	purego.RegisterLibFunc(&v.voicevoxErrorResultToMessage, v.core, "voicevox_error_result_to_message")
	purego.RegisterLibFunc(&v.voicevoxGetVersion, v.core, "voicevox_get_version")
	purego.RegisterLibFunc(&v.voicevoxJsonFree, v.core, "voicevox_json_free")
	purego.RegisterLibFunc(&v.voicevoxWavFree, v.core, "voicevox_wav_free")
}

// シンボルが存在する場合のみ紐付けを行う
func (v *Voicevox) registerOptional(fptr any, name string) {
	if _, err := purego.Dlsym(v.core, name); err != nil {
		return
	}
	purego.RegisterLibFunc(fptr, v.core, name)
}
