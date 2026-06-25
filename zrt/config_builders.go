package zrt

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
)

// envKeyMap maps environment variables to the credential slots they populate.
var envKeyMap = map[string][]string{
	"DEEPGRAM_API_KEY":          {"deepgram"},
	"GOOGLE_API_KEY":            {"gemini", "google", "google_stt"},
	"OPENAI_API_KEY":            {"openai", "openai_stt"},
	"ANTHROPIC_API_KEY":         {"anthropic", "claude"},
	"CARTESIA_API_KEY":          {"cartesia"},
	"ELEVENLABS_API_KEY":        {"elevenlabs"},
	"GROQ_API_KEY":              {"groq"},
	"CEREBRAS_API_KEY":          {"cerebras"},
	"XAI_API_KEY":               {"xai", "grok"},
	"ASSEMBLYAI_API_KEY":        {"assemblyai"},
	"AZURE_SPEECH_KEY":          {"azure"},
	"AZURE_REGION":              {"azure_region"},
	"AZURE_VOICE_LIVE_API_KEY":  {"azure_voice_live"},
	"AZURE_VOICE_LIVE_ENDPOINT": {"azure_voice_live_endpoint"},
	"NVIDIA_API_KEY":            {"nvidia", "nvidia_riva"},
	"GLADIA_API_KEY":            {"gladia"},
	"SARVAMAI_API_KEY":          {"sarvamai", "sarvam"},
	"COMETAPI_API_KEY":          {"cometapi", "comet"},
	"COMETAPI_BASE_URL":         {"cometapi_base_url"},
}

// knobMap lists the per-provider tuning knobs serialized into credentials.
var knobMap = map[string][]string{
	"cartesia":   {"model", "language", "speed", "volume", "emotion", "max_buffer_delay_ms", "pronunciation_dict_id", "enable_word_timestamps"},
	"deepgram":   {"smart_format", "punctuate", "filler_words", "profanity_filter", "numerals", "tag", "keywords", "keyterm", "redact", "interim_results", "diarize", "base_url"},
	"silero":     {"smoothing_factor", "min_volume"},
	"elevenlabs": {"model", "stability", "similarity_boost", "style", "use_speaker_boost", "apply_text_normalization", "enable_word_timestamps"},
	"anthropic":  {"thinking_budget"},
	"gemini":     {"thinking_budget", "include_thoughts", "top_p", "top_k", "presence_penalty", "frequency_penalty", "seed"},
	"openai":     {"top_p", "frequency_penalty", "presence_penalty", "seed", "response_format", "tool_choice", "parallel_tool_calls"},
	"sarvamai":   {"model", "language", "streaming", "pitch", "pace", "loudness", "temperature", "preprocessing", "bitrate"},
	"google":     {"language", "voice", "speaking_rate", "pitch"},
}

// sttKnobMap lists the per-provider STT-specific tuning knobs.
var sttKnobMap = map[string][]string{
	"sarvamai":   {"mode", "translation", "prompt", "high_vad_sensitivity", "flush_signals", "input_sample_rate", "output_sample_rate"},
	"google_stt": {"punctuate", "profanity_filter"},
}

// DefaultVoiceSuffix is appended to instructions unless overridden.
const DefaultVoiceSuffix = "\n\n[Voice conversation rules]\n- Keep responses concise: 1-3 sentences max unless asked for detail.\n- Use natural conversational language, not formal writing.\n- Never use markdown, bullet points, or formatting — this is spoken audio.\n"

// ---- value coercion helpers ----

// floatStr formats a float so the result always contains a decimal point.
func floatStr(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eEnN") {
		s += ".0"
	}
	return s
}

// serializeKnob applies the general knob serialization rules. ok=false means
// the value should be skipped (nil / empty string).
func serializeKnob(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", false
	case string:
		if x == "" {
			return "", false
		}
		return x, true
	case bool:
		if x {
			return "true", true
		}
		return "false", true
	case int:
		return strconv.Itoa(x), true
	case int32:
		return strconv.FormatInt(int64(x), 10), true
	case int64:
		return strconv.FormatInt(x, 10), true
	case uint32:
		return strconv.FormatUint(uint64(x), 10), true
	case float32:
		return floatStr(float64(x)), true
	case float64:
		return floatStr(x), true
	case []string:
		return strings.Join(x, ","), true
	case []float64, []int, []any, []map[string]any:
		b, _ := json.Marshal(x)
		return string(b), true
	case map[string]any, map[string]string:
		b, _ := json.Marshal(x)
		return string(b), true
	default:
		return fmt.Sprintf("%v", x), true
	}
}

// ---- provider config builders ----

func buildSTTConfig(stt STT) *pb.STTProviderConfig {
	if stt == nil {
		return &pb.STTProviderConfig{}
	}
	c := stt.STTConfig()
	return &pb.STTProviderConfig{
		Provider:      c.Provider,
		Model:         c.Model,
		Language:      c.Language,
		EndpointingMs: cmp.Or(c.EndpointingMs, 50),
	}
}

func buildLLMConfig(llm LLM) *pb.LLMProviderConfig {
	if llm == nil {
		return &pb.LLMProviderConfig{}
	}
	c := llm.LLMConfig()
	cfg := &pb.LLMProviderConfig{
		Provider:        c.Provider,
		Model:           c.Model,
		Temperature:     c.Temperature,
		MaxOutputTokens: c.MaxOutputTokens,
	}
	if llm.ProviderName() == "gemini" {
		if ge, ok := llm.(GeminiExtrasProvider); ok {
			if extras := buildGeminiLLMExtras(ge.GeminiLLMExtras()); extras != nil {
				cfg.GeminiExtras = extras
			}
		}
	}
	return cfg
}

func buildGeminiLLMExtras(e *GeminiLLMExtras) *pb.GeminiLLMExtras {
	if e == nil {
		return nil
	}
	hasVertex := e.Vertex != nil && e.Vertex.ProjectID != ""
	if len(e.SafetySettings) == 0 && e.ThinkingBudget == nil && !e.IncludeThoughts && !hasVertex {
		return nil
	}
	extras := &pb.GeminiLLMExtras{}
	if e.ThinkingBudget != nil {
		v := int32(*e.ThinkingBudget)
		extras.ThinkingBudget = &v
	}
	if e.IncludeThoughts {
		t := true
		extras.IncludeThoughts = &t
	}
	for _, s := range e.SafetySettings {
		if s.Category != "" && s.Threshold != "" {
			extras.SafetySettings = append(extras.SafetySettings, &pb.SafetySetting{Category: s.Category, Threshold: s.Threshold})
		}
	}
	if hasVertex {
		v := &pb.VertexAIConfig{ProjectId: e.Vertex.ProjectID}
		if e.Vertex.Location != "" {
			v.Location = e.Vertex.Location
		}
		if e.Vertex.ServiceAccountJSON != "" {
			v.ServiceAccountJson = e.Vertex.ServiceAccountJSON
		}
		if e.Vertex.ServiceAccountPath != "" {
			v.ServiceAccountPath = e.Vertex.ServiceAccountPath
		}
		extras.Vertex = v
	}
	return extras
}

func buildTTSConfig(tts TTS) *pb.TTSProviderConfig {
	if tts == nil {
		return &pb.TTSProviderConfig{}
	}
	c := tts.TTSConfig()
	cfg := &pb.TTSProviderConfig{Provider: c.Provider, Voice: c.Voice}
	if tts.ProviderName() == "cartesia" {
		if emb := tts.VoiceEmbedding(); len(emb) > 0 {
			ve := make([]float32, len(emb))
			for i, x := range emb {
				ve[i] = float32(x)
			}
			cfg.CartesiaExtras = &pb.CartesiaExtras{VoiceEmbedding: ve}
		}
	}
	return cfg
}

func buildVADConfig(vad VAD) *pb.VADProviderConfig {
	if vad == nil {
		return &pb.VADProviderConfig{}
	}
	c := vad.VADConfig()
	stop := cmp.Or(c.StopThreshold, 0.25)
	smoothing := cmp.Or(c.SmoothingFactor, 0.35)
	return &pb.VADProviderConfig{
		Threshold:          c.Threshold,
		MinSpeechMs:        uint32(c.MinSpeechDuration * 1000),
		MinSilenceMs:       uint32(c.MinSilenceDuration * 1000),
		StopThreshold:      stop,
		MinSpeechDuration:  c.MinSpeechDuration,
		MinSilenceDuration: c.MinSilenceDuration,
		PaddingDuration:    c.PaddingDuration,
		MaxBufferedSpeech:  c.MaxBufferedSpeech,
		ForceCpu:           c.ForceCPU,
		SmoothingFactor:    smoothing,
	}
}

var interruptModeMap = map[string]string{"VAD_ONLY": "vad_only", "STT_ONLY": "stt_only", "HYBRID": "hybrid"}

func buildInterruptConfig(ic *InterruptConfig) *pb.InterruptConfig {
	if ic == nil {
		return &pb.InterruptConfig{Mode: "hybrid", MinDurationMs: 500, MinWords: 2, CooldownMs: 500}
	}
	ic.normalize()
	mode := cmp.Or(interruptModeMap[ic.Mode], "hybrid")
	pauseMS := cmp.Or(ic.FalseInterruptPauseDurationMS, int(ic.FalseInterruptPauseDuration*1000))
	return &pb.InterruptConfig{
		Mode:                          mode,
		MinDurationMs:                 uint32(ic.InterruptMinDuration * 1000),
		MinWords:                      uint32(ic.InterruptMinWords),
		CooldownMs:                    500,
		FalseInterruptPauseDurationMs: uint32(pauseMS),
		ResumeOnFalseInterrupt:        ic.ResumeOnFalseInterrupt,
		InterruptFadeDurationMs:       uint32(ic.InterruptFadeDurationMS),
	}
}

func buildEOUConfig(eou *EOUConfig, turnDetector EOU) *pb.EOUConfig {
	var minWait, maxWait uint32 = 50, 400
	if eou != nil {
		w := eou.MinMaxSpeechWaitTimeout
		if len(w) > 0 {
			minWait = uint32(w[0] * 1000)
		}
		if len(w) > 1 {
			maxWait = uint32(w[1] * 1000)
		}
	}
	var provider, modelID, host, authToken string
	threshold := float32(0.9)
	if turnDetector != nil {
		provider = turnDetector.ProviderName()
		tc := turnDetector.TurnConfig()
		modelID = tc.ModelID
		host = tc.Host
		authToken = tc.AuthToken
		if tc.HasThreshold && tc.Threshold > 0 {
			threshold = tc.Threshold
		}
	}
	mode := ""
	if eou != nil {
		mode = strings.ToLower(eou.Mode)
	}
	return &pb.EOUConfig{
		Threshold: threshold,
		MinWaitMs: minWait,
		MaxWaitMs: maxWait,
		Provider:  provider,
		ModelId:   modelID,
		Host:      host,
		AuthToken: authToken,
		Mode:      mode,
	}
}

func buildToolSchemas(tools []*FunctionTool) []*pb.ToolSchemaProto {
	var out []*pb.ToolSchemaProto
	for _, t := range tools {
		if !IsFunctionTool(t) {
			continue
		}
		schema := t.Info.ParametersSchema
		if schema == nil {
			schema = map[string]any{}
		}
		b, _ := json.Marshal(schema)
		out = append(out, &pb.ToolSchemaProto{
			Name:                 t.Info.Name,
			Description:          t.Info.Description,
			ParametersJsonSchema: string(b),
		})
	}
	return out
}
func resolveContextWindow(agent Agent, p *Pipeline) *ContextWindow {
	if agent != nil {
		if cw := agent.base().contextWindow(); cw != nil {
			return cw
		}
	}
	if p != nil {
		return p.ContextWindow
	}
	return nil
}

func buildContextWindowConfig(cw *ContextWindow) *pb.ContextWindowConfig {
	ctx := &pb.ContextWindowConfig{}
	if cw == nil {
		return ctx
	}
	if cw.MaxTokens != 0 {
		ctx.MaxTokens = uint32(cw.MaxTokens)
	}
	if cw.MaxContextItems != 0 {
		ctx.MaxContextItems = uint32(cw.MaxContextItems)
	}
	ctx.KeepRecentTurns = uint32(cw.KeepRecentTurns)
	ctx.MaxToolCallsPerTurn = uint32(cw.MaxToolCallsPerTurn)
	if cw.SummaryLLM != nil {
		ctx.SummaryLlm = buildLLMConfig(cw.SummaryLLM)
	}
	return ctx
}

// inferenceMarked is implemented by components that may be dedicated-inference.
type inferenceMarked interface {
	ProviderName() string
	InferenceInfo() InferenceInfo
}

func buildCredentials(p *Pipeline, sessionOptions map[string]string, agent Agent) map[string]string {
	creds := map[string]string{}
	for envVar, keys := range envKeyMap {
		if v := os.Getenv(envVar); v != "" {
			for _, k := range keys {
				creds[k] = v
			}
		}
	}
	// Provider API keys.
	for _, prov := range []Provider{as[Provider](p.STT), as[Provider](p.LLM), as[Provider](p.TTS)} {
		if prov == nil {
			continue
		}
		if key := prov.APIKey(); key != "" {
			if name := prov.ProviderName(); name != "" {
				creds[name] = key
			}
		}
	}
	if cw := resolveContextWindow(agent, p); cw != nil && cw.SummaryLLM != nil {
		sl := cw.SummaryLLM
		if key := sl.APIKey(); key != "" {
			if name := sl.ProviderName(); name != "" {
				creds[name] = key
			}
		}
	}
	// General KNOB_MAP over stt, llm, tts, vad.
	for _, prov := range []Provider{as[Provider](p.STT), as[Provider](p.LLM), as[Provider](p.TTS), as[Provider](p.VAD)} {
		if prov == nil {
			continue
		}
		name := prov.ProviderName()
		knobs := prov.Knobs()
		for _, knob := range knobMap[name] {
			if v, ok := knobs[knob]; ok {
				if s, keep := serializeKnob(v); keep {
					creds[name+"_"+knob] = s
				}
			}
		}
	}
	// STT_KNOB_MAP.
	if p.STT != nil {
		name := p.STT.ProviderName()
		knobs := p.STT.Knobs()
		for _, knob := range sttKnobMap[name] {
			if v, ok := knobs[knob]; ok {
				if s, keep := serializeKnob(v); keep {
					creds[name+"_stt_"+knob] = s
				}
			}
		}
	}
	// Denoise credentials.
	if p.Denoise != nil {
		denoiseProvider := p.Denoise.ProviderName()
		if denoiseProvider != "" {
			token := p.Denoise.GatewayToken
			if token == "" {
				token = os.Getenv("ZRT_AUTH_TOKEN")
			}
			if token != "" {
				creds["denoise_gateway_token"] = token
			}
			if p.Denoise.ModelID != "" {
				creds["denoise_model"] = p.Denoise.ModelID
			}
			if p.Denoise.hasModelSampleRate {
				creds["denoise_model_sample_rate"] = strconv.Itoa(p.Denoise.ModelSampleRate)
			}
			if p.Denoise.hasChunkMS {
				creds["denoise_chunk_ms"] = strconv.Itoa(p.Denoise.ChunkMS)
			}
			if p.Denoise.BaseURL != "" {
				creds["denoise_base_url"] = p.Denoise.BaseURL
			}
		}
	}
	// Dedicated-inference markers.
	for _, m := range []inferenceMarked{as[inferenceMarked](p.STT), as[inferenceMarked](p.LLM), as[inferenceMarked](p.TTS), as[inferenceMarked](p.TurnDetector)} {
		if m == nil {
			continue
		}
		info := m.InferenceInfo()
		if !info.IsInference {
			continue
		}
		name := m.ProviderName()
		if name == "" {
			continue
		}
		creds[name+"_inference"] = "true"
		if info.BaseURL != "" {
			if _, exists := creds[name+"_base_url"]; !exists {
				creds[name+"_base_url"] = info.BaseURL
			}
		}
		if info.Location != "" {
			creds[name+"_location"] = info.Location
		}
	}
	for k, v := range sessionOptions {
		if v != "" {
			creds[k] = v
		}
	}
	return creds
}

func buildDenoiseConfig(d *Denoise) *pb.DenoiseConfig {
	if d == nil {
		return &pb.DenoiseConfig{Enabled: false, Provider: ""}
	}
	return &pb.DenoiseConfig{Enabled: true, Provider: d.ProviderName()}
}

func buildVoicemailConfig(v *VoiceMailDetector) *pb.VoicemailConfig {
	if v == nil {
		return &pb.VoicemailConfig{Enabled: false}
	}
	duration := cmp.Or(v.Duration, 2.0)
	threshold := cmp.Or(v.DetectionThreshold, 1.0)
	return &pb.VoicemailConfig{
		Enabled:             true,
		DetectionThreshold:  float32(threshold),
		MaxDetectionSeconds: uint32(roundFloat(duration)),
		CustomPrompt:        v.CustomPrompt,
		AutoHangup:          v.AutoHangup,
	}
}

func roundFloat(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

var fillerWords = []string{"okay", "ok", "yeah", "yes", "right", "sure", "hmm", "uh huh", "mhm", "huh", "what", "wait", "sorry", "pardon", "repeat", "again", "say again", "say that again", "what was that", "what did you say", "come again", "kya", "haan", "han", "kya kaha", "kya bola", "kya kaha aapne", "phir se", "phir bolo", "phir boliye", "dobara", "dobara bolo", "dobara boliye", "dobara kaho", "samjha nahi", "samjhi nahi", "samajh nahi aaya", "maaf kijiye", "maaf karna"}

var verbalFillers = []string{"Hmm,", "Let me think.", "Sure,", "Okay,"}

func buildCascadeConfig(p *Pipeline) *pb.CascadeConfig {
	sttSlot := p.STT
	ttsSlot := p.TTS
	// Custom STT/TTS hook placeholders.
	var sttCfg *pb.STTProviderConfig
	if sttSlot == nil && p.Hooks != nil && p.Hooks.hasSTTStreamHook() {
		sttCfg = &pb.STTProviderConfig{Provider: "custom"}
	} else {
		sttCfg = buildSTTConfig(sttSlot)
	}
	var ttsCfg *pb.TTSProviderConfig
	if ttsSlot == nil && p.Hooks != nil && p.Hooks.hasTTSStreamHook() {
		ttsCfg = &pb.TTSProviderConfig{Provider: "custom"}
	} else {
		ttsCfg = buildTTSConfig(ttsSlot)
	}

	minClause := uint32(cmp.Or(p.SentenceMinClauseLen, 5))
	firstClause := uint32(cmp.Or(p.SentenceFirstClauseLen, 3))
	minSentence := uint32(cmp.Or(p.SentenceMinSentenceLen, 3))

	return &pb.CascadeConfig{
		Stt:       sttCfg,
		Llm:       buildLLMConfig(as[LLM](p.LLM)),
		Tts:       ttsCfg,
		Vad:       buildVADConfig(p.VAD),
		Interrupt: buildInterruptConfig(p.InterruptConfig),
		Eou:       buildEOUConfig(p.EOUConfig, p.TurnDetector),
		Denoise:   buildDenoiseConfig(p.Denoise),
		Voicemail: buildVoicemailConfig(p.VoiceMailDetector),
		Filler: &pb.FillerConfig{
			FilterFillerWords:    true,
			FillerWords:          fillerWords,
			EnableVerbalFillers:  false,
			VerbalFillers:        verbalFillers,
			VerbalFillerMinWords: 3,
		},
		Audio: &pb.AudioConfig{
			TtsSampleRate:   48000,
			PlaybackGraceMs: uint32(p.PlaybackGraceMS),
		},
		SentenceBuffer: &pb.SentenceBufferConfig{
			MinClauseLen:   minClause,
			FirstClauseLen: firstClause,
			MinSentenceLen: minSentence,
		},
		OrchestratorTiming:   &pb.OrchestratorTimingConfig{PollIntervalMs: 10, MaxDrainSecs: 10, TtsDrainGraceMs: 500},
		SttFilterPatterns:    slices.Clone(p.STTFilterPatterns),
		SttWordSubstitutions: maps.Clone(p.STTWordSubstitutions),
	}
}

func llmIsRealtime(llm LLMLike) bool {
	if llm == nil {
		return false
	}
	rm, ok := llm.(RealtimeModel)
	return ok && rm.IsRealtimeModel()
}

func detectPipelineMode(p *Pipeline) string {
	rt := p.RealtimeConfig
	explicit := ""
	if rt != nil {
		explicit = rt.Mode
	}
	switch explicit {
	case "full_s2s", "s2s", "realtime":
		return "realtime"
	case "hybrid_tts":
		return "hybrid_tts"
	case "hybrid_stt":
		return "hybrid_stt"
	case "llm_only":
		return "llm_only"
	}
	llmRealtime := llmIsRealtime(p.LLM)
	hasSTT := p.STT != nil
	hasTTS := p.TTS != nil
	hasLLM := p.LLM != nil
	switch {
	case llmRealtime && hasSTT && hasTTS:
		return "llm_only"
	case llmRealtime && hasTTS:
		return "hybrid_tts"
	case llmRealtime && hasSTT:
		return "hybrid_stt"
	case llmRealtime:
		return "realtime"
	case hasSTT && hasLLM && hasTTS:
		return "cascade"
	case hasSTT && hasLLM:
		return "stt_llm_only"
	case hasLLM && hasTTS:
		return "llm_tts_only"
	case hasSTT && hasTTS:
		return "stt_tts_only"
	case hasSTT:
		return "stt_only"
	case hasTTS:
		return "tts_only"
	case hasLLM:
		return "llm_only"
	}
	if rt != nil && explicit == "" {
		return "realtime"
	}
	return "cascade"
}

func buildRealtimeProviderConfig(p *Pipeline) *pb.RealtimeProviderConfig {
	var info RealtimeInfo
	provider := ""
	if rm, ok := p.LLM.(RealtimeModel); ok {
		info = rm.RealtimeInfo()
		provider = cmp.Or(info.Provider, rm.ProviderName())
	}
	params := maps.Clone(info.Params)
	var modalities []string
	if p.RealtimeConfig != nil && len(p.RealtimeConfig.ResponseModalities) > 0 {
		modalities = slices.Clone(p.RealtimeConfig.ResponseModalities)
		if p.RealtimeConfig.Mode == "llm_only" {
			modalities = []string{"TEXT"}
		}
	}
	if len(modalities) == 0 {
		modalities = slices.Clone(info.ResponseModalities)
	}
	detected := detectPipelineMode(p)
	protoMode := "full_s2s"
	switch detected {
	case "hybrid_stt", "hybrid_tts", "llm_only":
		protoMode = detected
	}
	if detected == "llm_only" {
		modalities = []string{"TEXT"}
	}
	cfg := &pb.RealtimeProviderConfig{
		Provider:           provider,
		Model:              info.Model,
		Voice:              info.Voice,
		RealtimeMode:       protoMode,
		ResponseModalities: modalities,
		Params:             params,
	}
	if (provider == "gemini" || provider == "gemini_live") && info.Vertex != nil && info.Vertex.ProjectID != "" {
		loc := cmp.Or(info.Vertex.Location, "us-central1")
		vx := &pb.VertexAIConfig{ProjectId: info.Vertex.ProjectID, Location: loc}
		if info.Vertex.ServiceAccountJSON != "" {
			vx.ServiceAccountJson = info.Vertex.ServiceAccountJSON
		} else if info.Vertex.ServiceAccountPath != "" {
			vx.ServiceAccountPath = info.Vertex.ServiceAccountPath
		}
		cfg.GeminiLiveExtras = &pb.GeminiLiveExtras{Vertex: vx}
	}
	return cfg
}

func buildPipelineConfig(p *Pipeline) *pb.PipelineConfig {
	mode := detectPipelineMode(p)
	switch mode {
	case "realtime":
		return &pb.PipelineConfig{Mode: "realtime", Realtime: buildRealtimeProviderConfig(p)}
	case "hybrid_tts":
		return &pb.PipelineConfig{Mode: "hybrid_tts", Cascade: buildCascadeConfig(p), Realtime: buildRealtimeProviderConfig(p)}
	case "hybrid_stt":
		return &pb.PipelineConfig{Mode: "hybrid_stt", Cascade: buildCascadeConfig(p), Realtime: buildRealtimeProviderConfig(p)}
	case "llm_only":
		if llmIsRealtime(p.LLM) {
			return &pb.PipelineConfig{Mode: "llm_only", Cascade: buildCascadeConfig(p), Realtime: buildRealtimeProviderConfig(p)}
		}
		return &pb.PipelineConfig{Mode: "llm_only", Cascade: buildCascadeConfig(p)}
	case "stt_only", "tts_only", "stt_tts_only", "stt_llm_only", "llm_tts_only", "partial_cascading", "text_only":
		return &pb.PipelineConfig{Mode: mode, Cascade: buildCascadeConfig(p)}
	}
	return &pb.PipelineConfig{Mode: "cascade", Cascade: buildCascadeConfig(p)}
}

func buildMCPServerConfigs(a *BaseAgent) []*pb.MCPServerConfig {
	var out []*pb.MCPServerConfig
	for _, srv := range a.mcpServers {
		switch srv.mcpType() {
		case "stdio":
			cmd, args, env := srv.mcpStdio()
			out = append(out, &pb.MCPServerConfig{Type: "stdio", Command: cmd, Args: args, Env: env})
		case "http":
			url, headers := srv.mcpHTTP()
			out = append(out, &pb.MCPServerConfig{Type: "http", Url: url, Env: maps.Clone(headers)})
		default:
			out = append(out, &pb.MCPServerConfig{Type: srv.mcpType()})
		}
	}
	return out
}

func buildKnowledgeBaseConfig(a *BaseAgent) *pb.KnowledgeBaseConfig {
	kb := a.knowledgeBase
	if kb == nil {
		return &pb.KnowledgeBaseConfig{Enabled: false}
	}
	cfg := kb.Config
	if cfg == nil {
		return &pb.KnowledgeBaseConfig{Enabled: true}
	}
	return &pb.KnowledgeBaseConfig{
		Enabled:   true,
		Provider:  cmp.Or(cfg.Provider, "custom"),
		IndexName: cfg.IndexName,
		TopK:      uint32(cmp.Or(cfg.TopK, 5)),
		MinScore:  float32(cmp.Or(cfg.MinScore, 0.7)),
		Params:    maps.Clone(cfg.Params),
	}
}

func buildAgentConfig(agent Agent, p *Pipeline) *pb.AgentConfig {
	a := agent.base()
	userGreeting := a.greeting
	var suffixToSend string
	if a.voiceSuffix == nil {
		if a.appendVoiceSuffix {
			suffixToSend = DefaultVoiceSuffix
		}
	} else {
		suffixToSend = *a.voiceSuffix
	}
	var altProtos []*pb.NamedAgentConfig
	seenAlt := map[string]bool{}
	for _, ag := range a.handoffAgents {
		if ag == nil {
			continue
		}
		hb := ag.base()
		if hb.id == "" || seenAlt[hb.id] {
			continue
		}
		seenAlt[hb.id] = true
		altProtos = append(altProtos, &pb.NamedAgentConfig{
			AgentId:      hb.id,
			Instructions: hb.instructions,
			Tools:        buildToolSchemas(hb.tools),
			Greeting:     hb.greeting,
		})
	}
	for _, alt := range a.alternates {
		if alt == nil || alt.AgentID == "" || seenAlt[alt.AgentID] {
			continue
		}
		seenAlt[alt.AgentID] = true
		altProtos = append(altProtos, &pb.NamedAgentConfig{
			AgentId:      alt.AgentID,
			Instructions: alt.Instructions,
			Tools:        buildToolSchemas(alt.Tools),
			Greeting:     alt.Greeting,
		})
	}
	var registeredHooks []string
	if p != nil && p.Hooks != nil {
		registeredHooks = p.Hooks.registeredNames()
	}
	autoEnable := map[string]bool{}
	for _, n := range registeredHooks {
		if hookNamesAutoEnable[n] {
			autoEnable[n] = true
		}
	}
	autoBeforeLLM := autoEnable["llm"] || autoEnable["llm_messages"]
	autoLLMStream := autoEnable["llm_stream"]
	llmStreamEnabled := autoLLMStream || a.llmStreamHookEnabled || (p != nil && p.LLMStreamHookEnabled)
	timeoutMS := cmp.Or(a.llmStreamHookTimeoutMS, 100)
	contextWindow := resolveContextWindow(agent, p)
	return &pb.AgentConfig{
		AgentId:                  a.id,
		Instructions:             a.instructions,
		Tools:                    buildToolSchemas(a.tools),
		RegisteredHooks:          registeredHooks,
		BeforeLlmHook:            autoBeforeLLM || a.beforeLLMHook,
		ContextWindow:            buildContextWindowConfig(contextWindow),
		Greeting:                 userGreeting,
		GreetingNonInterruptible: a.greetingNonInterruptible,
		AppendVoiceSuffix:        a.appendVoiceSuffix,
		VoiceSuffix:              suffixToSend,
		LlmStreamHookEnabled:     llmStreamEnabled,
		LlmStreamHookTimeoutMs:   uint32(timeoutMS),
		Alternates:               altProtos,
		McpServers:               buildMCPServerConfigs(a),
		KnowledgeBase:            buildKnowledgeBaseConfig(a),
		InheritContext:           a.inheritContext,
	}
}

var (
	recordingFormatMap        = map[string]pb.RecordingFormat{"wav": 0, "ogg_opus": 1, "mp3": 2, "flac": 3}
	recordingChannelMap       = map[string]pb.RecordingChannelMode{"mixed": 0, "dual_channel": 1}
	recordingTranscriptFmtMap = map[string]pb.RecordingTranscriptFormat{"json": 0, "srt": 1, "vtt": 2}
)

func buildRecordingConfig(rec *RecordingConfig) *pb.RecordingConfig {
	if rec == nil {
		return &pb.RecordingConfig{Enabled: false}
	}
	format := pb.RecordingFormat(1)
	if v, ok := recordingFormatMap[rec.Format]; ok {
		format = v
	}
	channel := pb.RecordingChannelMode(1)
	if v, ok := recordingChannelMap[rec.ChannelMode]; ok {
		channel = v
	}
	cfg := &pb.RecordingConfig{
		Enabled:            rec.Enabled,
		AutoStart:          rec.AutoStart,
		Format:             format,
		ChannelMode:        channel,
		SampleRate:         uint32(rec.SampleRate),
		BitrateKbps:        uint32(rec.BitrateKbps),
		MaxDurationSeconds: uint32(rec.MaxDurationSeconds),
		MaxFileSizeMb:      uint32(rec.MaxFileSizeMB),
		RecordingBeep:      rec.RecordingBeep,
		RedactDtmf:         rec.RedactDTMF,
		CustomMetadata:     maps.Clone(rec.CustomMetadata),
		RecordingName:      rec.RecordingName,
		RecordingGroup:     rec.RecordingGroup,
		ApplyDenoise:       rec.ApplyDenoise,
		NormalizeAudio:     rec.NormalizeAudio,
		TrimSilence:        rec.TrimSilence,
		RecordVideo:        rec.RecordVideo,
		RecordScreenShare:  rec.RecordScreenShare,
		RecordScreenAudio:  rec.RecordScreenAudio,
	}
	if rec.Storage != nil {
		s := rec.Storage
		cfg.Storage = &pb.RecordingStorageConfig{Storage: &pb.RecordingStorageConfig_S3{S3: &pb.S3StorageConfig{
			Bucket:               s.Bucket,
			Region:               s.Region,
			Prefix:               s.Prefix,
			AccessKeyId:          s.AccessKeyID,
			SecretAccessKey:      s.SecretAccessKey,
			SessionToken:         s.SessionToken,
			EndpointUrl:          s.EndpointURL,
			StorageClass:         s.StorageClass,
			ServerSideEncryption: s.ServerSideEncryption,
			KmsKeyId:             s.KMSKeyID,
			Acl:                  s.ACL,
			MultipartUpload:      s.MultipartUpload,
			MultipartPartSizeMb:  uint32(s.MultipartPartSizeMB),
			UploadTimeoutSeconds: uint32(s.UploadTimeoutSeconds),
			MaxRetryAttempts:     uint32(s.MaxRetryAttempts),
			Tags:                 maps.Clone(s.Tags),
			UserMetadata:         maps.Clone(s.UserMetadata),
			ContentTypeOverride:  s.ContentTypeOverride,
		}}}
	}
	if rec.Transcript != nil {
		t := rec.Transcript
		cfg.Transcript = &pb.RecordingTranscriptConfig{
			Enabled:               t.Enabled,
			Format:                recordingTranscriptFmtMap[t.Format],
			IncludeWordTimestamps: t.IncludeWordTimestamps,
			IncludeConfidence:     t.IncludeConfidence,
			SpeakerLabels:         t.SpeakerLabels,
			Language:              t.Language,
		}
	}
	return cfg
}

func buildSessionConfig(p *Pipeline, agent Agent, room roomConfigData, recording *RecordingConfig, sessionOptions map[string]string, sessionID string) *pb.SessionConfig {
	return &pb.SessionConfig{
		SessionId: sessionID,
		Pipeline:  buildPipelineConfig(p),
		Agent:     buildAgentConfig(agent, p),
		Room: &pb.RoomConfig{
			RoomId:                      room.RoomID,
			AuthToken:                   room.AuthToken,
			AgentName:                   cmp.Or(room.AgentName, "Agent"),
			AutoEndSession:              room.AutoEndSession,
			SessionTimeoutSeconds:       uint32(room.SessionTimeoutSeconds),
			NoParticipantTimeoutSeconds: uint32(room.NoParticipantTimeoutSeconds),
			AudioListenerEnabled:        room.AudioListenerEnabled,
			AgentParticipantId:          room.AgentParticipantID,
			Playground:                  room.Playground,
			Vision:                      room.Vision,
			RecordingEnabled:            room.RecordingEnabled,
			BackgroundAudioEnabled:      room.BackgroundAudioEnabled,
			SendLogsToDashboard:         true,
		},
		Credentials:   &pb.CredentialsConfig{ProviderKeys: buildCredentials(p, sessionOptions, agent)},
		Limits:        buildSessionLimits(agent, p),
		ClientVersion: clientVersionInfo(),
		Recording:     buildRecordingConfig(recording),
	}
}

func buildSessionLimits(agent Agent, p *Pipeline) *pb.SessionLimits {
	inactivity := uint32(300)
	if p != nil && p.InactivityTimeoutSeconds != nil && *p.InactivityTimeoutSeconds > 0 {
		inactivity = uint32(*p.InactivityTimeoutSeconds)
	}
	limits := &pb.SessionLimits{InactivityTimeoutSeconds: inactivity}
	if agent != nil {
		if d := agent.base().maxSessionDurationSeconds; d != nil && *d > 0 {
			limits.MaxSessionDurationSeconds = uint32(*d)
		}
	}
	return limits
}

func buildAgentRegistration(agent Agent, p *Pipeline, agentKind string, maxConcurrent int, authToken string, labels map[string]string, defaultRecording *RecordingConfig, sessionOptions map[string]string) *pb.AgentRegistration {
	return &pb.AgentRegistration{
		AgentKind:             agentKind,
		Agent:                 buildAgentConfig(agent, p),
		Pipeline:              buildPipelineConfig(p),
		Credentials:           &pb.CredentialsConfig{ProviderKeys: buildCredentials(p, sessionOptions, agent)},
		DefaultRecording:      buildRecordingConfig(defaultRecording),
		MaxConcurrentSessions: uint32(maxConcurrent),
		Labels:                labels,
		ClientVersion:         clientVersionInfo(),
		AuthToken:             authToken,
	}
}

// ---- type-assertion helper ----

// as returns v viewed as interface T, or the zero value of T if v is nil or
// does not implement T. Type-asserting a nil interface yields the zero value,
// so no explicit nil check is needed.
func as[T any](v any) T {
	t, _ := v.(T)
	return t
}
