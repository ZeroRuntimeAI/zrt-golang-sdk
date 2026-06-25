package zrt

import (
	"context"
	"sync"
)

// DefaultSTTFilterPatterns are the default transcript filter regexes.
var DefaultSTTFilterPatterns = []string{
	`\b(call|your call).*(being|now).*record(ed|ing)\b`,
	`\b(call recording).*(ended|stopped|no longer)\b`,
}

// ---- custom STT/TTS streaming payloads ----

// CustomSTTAudioChunk is raw audio handed to a custom STT hook.
type CustomSTTAudioChunk struct {
	PCM            []byte
	SampleRate     uint32
	UtteranceID    string
	EndOfUtterance bool
}

// CustomSTTResult is a transcript produced by a custom STT hook.
type CustomSTTResult struct {
	UtteranceID string
	Text        string
	IsFinal     bool
	Confidence  float64
	Language    string
	StartTimeUS uint64
	EndTimeUS   uint64
}

// CustomTTSSynthesize is a synthesis request handed to a custom TTS hook.
type CustomTTSSynthesize struct {
	Text        string
	UtteranceID string
	Voice       string
}

// CustomTTSAudioChunk is synthesized audio produced by a custom TTS hook.
type CustomTTSAudioChunk struct {
	UtteranceID    string
	PCM            []byte
	SampleRate     uint32
	EndOfSynthesis bool
}

// BeforeLLMData is delivered to a before-LLM hook.
type BeforeLLMData struct {
	Messages   []any
	TokenCount uint32
	TurnNumber uint32
}

// BeforeLLMResult is returned by a before-LLM hook to mutate or skip a turn.
type BeforeLLMResult struct {
	// Messages, if non-nil, replaces the LLM input messages.
	Messages []any
	// SkipTurn, if true, skips the LLM turn entirely.
	SkipTurn bool
}

// CustomSTTHook consumes audio chunks and returns a channel of transcripts.
// The returned channel must be closed when the hook is done.
type CustomSTTHook func(audio <-chan CustomSTTAudioChunk) <-chan CustomSTTResult

// CustomTTSHook consumes synthesis requests and returns a channel of audio.
type CustomTTSHook func(requests <-chan CustomTTSSynthesize) <-chan CustomTTSAudioChunk

// LLMStreamHook reviews each LLM token, returning a replacement (empty = keep)
// and whether to drop the token.
type LLMStreamHook func(text string, tokenID uint64) (replacement string, drop bool)

// STTHookData is the final transcript the runtime offers to the hook
// before it reaches the LLM (hook mode: real server-side STT + a hook).
type STTHookData struct {
	Text       string
	Language   string
	IsFinal    bool
	TurnNumber uint32
}

// STTHookResult is what a STT hook returns. ModifiedText
// empty (or equal to the original) keeps the transcript; Drop skips the turn.
type STTHookResult struct {
	ModifiedText string
	Drop         bool
}

// STTHookFunc rewrites a final transcript before the LLM sees it.
type STTHookFunc func(STTHookData) *STTHookResult

// TTSHookData is a text segment the runtime offers to the hook before the
// real server-side TTS synthesizes it.
type TTSHookData struct {
	Text        string
	UtteranceID string
	Voice       string
}

// TTSHookResult is what a TTS hook returns. ModifiedText empty
// (or equal to the original) keeps the text; Drop skips synthesizing the segment.
type TTSHookResult struct {
	ModifiedText string
	Drop         bool
}

// TTSHookFunc rewrites a text segment before server-side synthesis.
type TTSHookFunc func(TTSHookData) *TTSHookResult

// hookNamesAutoEnable lists hook names that are automatically enabled when registered.
var hookNamesAutoEnable = map[string]bool{"llm": true, "llm_stream": true, "llm_messages": true}

// PipelineHooks holds registered pipeline event hooks with typed registration.
type PipelineHooks struct {
	mu sync.Mutex

	customSTT CustomSTTHook
	customTTS CustomTTSHook
	llmStream LLMStreamHook

	beforeLLM func(BeforeLLMData) *BeforeLLMResult

	sttHook STTHookFunc
	ttsHook    TTSHookFunc

	generationStarted    []func(turnNumber uint32)
	generationComplete   []func(turnNumber uint32, wasInterrupted bool)
	generationChunk      []func(text string, metadata map[string]string)
	synthesisStarted     []func(text string)
	synthesisInterrupted []func(reason string)
	firstAudioByte       []func(ttfbMS, byteCount uint32)
	lastAudioByte        []func(durationSeconds float64)
	wordTiming           []func(word string, start, end float64, cumulative string)
	transcriptPreflight  []func(text string)
	eouDetected          []func(probability float64, waitMS uint32, text string)
	userTurnStart        []func(transcript string)
	userTurnEnd          []func()
	agentTurnStart       []func()
	agentTurnEnd         []func()
	llmCompleted         []func(text string, interrupted bool)
	llmTokenForReview    []func(text string, tokenID uint64) (replacement string, drop bool)
	errorHooks           []func(payload map[string]any)
	recordingStarted     []func(status map[string]any)
	recordingStopped     []func(status map[string]any)
	recordingFailed      []func(status map[string]any)
	visionFrame          []func(frame map[string]any)
	audioDelta           []func(frame map[string]any)
	metrics              map[string][]func(metrics map[string]any)

	registered map[string]bool
}

// metricsHooks returns the registered metrics callbacks for a component.
func (h *PipelineHooks) metricsHooks(component string) []func(map[string]any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.metrics[component]
}

func (h *PipelineHooks) mark(name string) {
	if h.registered == nil {
		h.registered = map[string]bool{}
	}
	h.registered[name] = true
}

// registeredNames returns the registered hook names.
func (h *PipelineHooks) registeredNames() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, 0, len(h.registered))
	for k := range h.registered {
		out = append(out, k)
	}
	return out
}

func (h *PipelineHooks) hasSTTStreamHook() bool { return h.customSTT != nil }
func (h *PipelineHooks) hasTTSStreamHook() bool { return h.customTTS != nil }

// Pipeline configures the voice stack the runtime executes.
type Pipeline struct {
	EventEmitter

	STT          STT
	LLM          LLMLike
	TTS          TTS
	VAD          VAD
	TurnDetector EOU
	Denoise      *Denoise

	EOUConfig         *EOUConfig
	InterruptConfig   *InterruptConfig
	ContextWindow     *ContextWindow
	VoiceMailDetector *VoiceMailDetector
	RealtimeConfig    *RealtimeConfig

	LLMStreamHookEnabled   bool
	LLMStreamHookTimeoutMS int
	PlaybackGraceMS        int

	STTFilterPatterns    []string
	STTWordSubstitutions map[string]string

	// Sentence buffer knobs (0 -> runtime defaults of 5/3/3).
	SentenceMinClauseLen   int
	SentenceFirstClauseLen int
	SentenceMinSentenceLen int

	Hooks *PipelineHooks
	agent Agent

	frameMu     sync.Mutex
	frameBuffer []map[string]any
}

// visionFrameBufferMax is the maximum number of vision frames retained in the buffer.
const visionFrameBufferMax = 5

// PipelineOptions configures a Pipeline.
type PipelineOptions struct {
	STT                    STT
	LLM                    LLMLike
	TTS                    TTS
	VAD                    VAD
	TurnDetector           EOU
	Denoise                *Denoise
	EOUConfig              *EOUConfig
	InterruptConfig        *InterruptConfig
	ContextWindow          *ContextWindow
	VoiceMailDetector      *VoiceMailDetector
	RealtimeConfig         *RealtimeConfig
	LLMStreamHookEnabled   bool
	LLMStreamHookTimeoutMS int
	PlaybackGraceMS        int
	// STTFilterPatterns overrides the default filter regexes. A nil slice uses
	// DefaultSTTFilterPatterns; pass an empty (non-nil) slice to disable.
	STTFilterPatterns    []string
	STTWordSubstitutions map[string]string
}

// NewPipeline builds a Pipeline from opts.
func NewPipeline(opts PipelineOptions) *Pipeline {
	patterns := opts.STTFilterPatterns
	if patterns == nil {
		patterns = append([]string(nil), DefaultSTTFilterPatterns...)
	}
	timeout := opts.LLMStreamHookTimeoutMS
	if timeout == 0 {
		timeout = 100
	}
	p := &Pipeline{
		STT:                    opts.STT,
		LLM:                    opts.LLM,
		TTS:                    opts.TTS,
		VAD:                    opts.VAD,
		TurnDetector:           opts.TurnDetector,
		Denoise:                opts.Denoise,
		EOUConfig:              opts.EOUConfig,
		InterruptConfig:        opts.InterruptConfig,
		ContextWindow:          opts.ContextWindow,
		VoiceMailDetector:      opts.VoiceMailDetector,
		RealtimeConfig:         opts.RealtimeConfig,
		LLMStreamHookEnabled:   opts.LLMStreamHookEnabled,
		LLMStreamHookTimeoutMS: timeout,
		PlaybackGraceMS:        opts.PlaybackGraceMS,
		STTFilterPatterns:      patterns,
		STTWordSubstitutions:   opts.STTWordSubstitutions,
		Hooks:                  &PipelineHooks{},
	}
	return p
}

func (p *Pipeline) setAgent(a Agent) { p.agent = a }

// ---- hook registration (typed, all on the Pipeline for convenience) ----

// OnCustomSTT registers a custom speech-to-text hook (sets the stt stream hook).
func (p *Pipeline) OnCustomSTT(h CustomSTTHook) { p.Hooks.customSTT = h; p.Hooks.mark("stt") }

// OnCustomTTS registers a custom text-to-speech hook (sets the tts stream hook).
func (p *Pipeline) OnCustomTTS(h CustomTTSHook) { p.Hooks.customTTS = h; p.Hooks.mark("tts") }

// OnLLMStream registers a per-token LLM review hook.
func (p *Pipeline) OnLLMStream(h LLMStreamHook) { p.Hooks.llmStream = h; p.Hooks.mark("llm_stream") }

// OnBeforeLLM registers a hook to mutate or skip an LLM turn before it runs.
func (p *Pipeline) OnBeforeLLM(h func(BeforeLLMData) *BeforeLLMResult) {
	p.Hooks.beforeLLM = h
	p.Hooks.mark("llm_messages")
}

// OnSTTHook registers a hook that rewrites the final transcript before
// it reaches the LLM. Requires a real server-side STT provider (hook mode).
func (p *Pipeline) OnSTTHook(h STTHookFunc) {
	p.Hooks.sttHook = h
	p.Hooks.mark("stt_hook")
}

// OnTTSHook registers a hook that rewrites a text segment before the real
// server-side TTS synthesizes it. Requires a real server-side TTS provider.
func (p *Pipeline) OnTTSHook(h TTSHookFunc) {
	p.Hooks.ttsHook = h
	p.Hooks.mark("tts_hook")
}

// metricsComponents is the set of components the runtime emits metrics for.
var metricsComponents = map[string]bool{"stt": true, "llm": true, "tts": true, "eou": true, "realtime": true}

// OnMetrics registers a component-level observability hook. component is one of
// "stt", "llm", "tts", "eou", "realtime"; the callback receives a per-turn
// metrics map (latency, TTFB, token counts) emitted by the runtime.
func (p *Pipeline) OnMetrics(component string, h func(metrics map[string]any)) {
	if !metricsComponents[component] {
		logger.Warnf("unknown metrics component %q", component)
		return
	}
	p.Hooks.mu.Lock()
	if p.Hooks.metrics == nil {
		p.Hooks.metrics = map[string][]func(map[string]any){}
	}
	p.Hooks.metrics[component] = append(p.Hooks.metrics[component], h)
	p.Hooks.mu.Unlock()
}

// OnLLMTokenForReview registers a fallback per-token review hook.
func (p *Pipeline) OnLLMTokenForReview(h func(text string, tokenID uint64) (string, bool)) {
	p.Hooks.llmTokenForReview = append(p.Hooks.llmTokenForReview, h)
	p.Hooks.mark("llm_token_for_review")
}

// OnLLMCompleted registers an observe hook fired when the LLM finishes a turn.
func (p *Pipeline) OnLLMCompleted(h func(text string, interrupted bool)) {
	p.Hooks.llmCompleted = append(p.Hooks.llmCompleted, h)
	p.Hooks.mark("llm")
}

// OnGenerationStarted registers a generation-started hook.
func (p *Pipeline) OnGenerationStarted(h func(turnNumber uint32)) {
	p.Hooks.generationStarted = append(p.Hooks.generationStarted, h)
	p.Hooks.mark("generation_started")
}

// OnGenerationComplete registers a generation-complete hook.
func (p *Pipeline) OnGenerationComplete(h func(turnNumber uint32, wasInterrupted bool)) {
	p.Hooks.generationComplete = append(p.Hooks.generationComplete, h)
	p.Hooks.mark("generation_complete")
}

// OnGenerationChunk registers a generation-chunk hook.
func (p *Pipeline) OnGenerationChunk(h func(text string, metadata map[string]string)) {
	p.Hooks.generationChunk = append(p.Hooks.generationChunk, h)
	p.Hooks.mark("generation_chunk")
}

// OnSynthesisStarted registers a synthesis-started hook.
func (p *Pipeline) OnSynthesisStarted(h func(text string)) {
	p.Hooks.synthesisStarted = append(p.Hooks.synthesisStarted, h)
	p.Hooks.mark("synthesis_started")
}

// OnSynthesisInterrupted registers a synthesis-interrupted hook.
func (p *Pipeline) OnSynthesisInterrupted(h func(reason string)) {
	p.Hooks.synthesisInterrupted = append(p.Hooks.synthesisInterrupted, h)
	p.Hooks.mark("synthesis_interrupted")
}

// OnFirstAudioByte registers a first-audio-byte hook.
func (p *Pipeline) OnFirstAudioByte(h func(ttfbMS, byteCount uint32)) {
	p.Hooks.firstAudioByte = append(p.Hooks.firstAudioByte, h)
	p.Hooks.mark("first_audio_byte")
}

// OnLastAudioByte registers a last-audio-byte hook.
func (p *Pipeline) OnLastAudioByte(h func(durationSeconds float64)) {
	p.Hooks.lastAudioByte = append(p.Hooks.lastAudioByte, h)
	p.Hooks.mark("last_audio_byte")
}

// OnWordTiming registers a word-timing hook.
func (p *Pipeline) OnWordTiming(h func(word string, start, end float64, cumulative string)) {
	p.Hooks.wordTiming = append(p.Hooks.wordTiming, h)
	p.Hooks.mark("word_timing")
}

// OnTranscriptPreflight registers a transcript-preflight hook.
func (p *Pipeline) OnTranscriptPreflight(h func(text string)) {
	p.Hooks.transcriptPreflight = append(p.Hooks.transcriptPreflight, h)
	p.Hooks.mark("transcript_preflight")
}

// OnEOUDetected registers an end-of-utterance hook.
func (p *Pipeline) OnEOUDetected(h func(probability float64, waitMS uint32, text string)) {
	p.Hooks.eouDetected = append(p.Hooks.eouDetected, h)
	p.Hooks.mark("eou_detected")
}

// OnUserTurnStart registers a user-turn-start hook.
func (p *Pipeline) OnUserTurnStart(h func(transcript string)) {
	p.Hooks.userTurnStart = append(p.Hooks.userTurnStart, h)
	p.Hooks.mark("user_turn_start")
}

// OnUserTurnEnd registers a user-turn-end hook.
func (p *Pipeline) OnUserTurnEnd(h func()) {
	p.Hooks.userTurnEnd = append(p.Hooks.userTurnEnd, h)
	p.Hooks.mark("user_turn_end")
}

// OnAgentTurnStart registers an agent-turn-start hook.
func (p *Pipeline) OnAgentTurnStart(h func()) {
	p.Hooks.agentTurnStart = append(p.Hooks.agentTurnStart, h)
	p.Hooks.mark("agent_turn_start")
}

// OnAgentTurnEnd registers an agent-turn-end hook.
func (p *Pipeline) OnAgentTurnEnd(h func()) {
	p.Hooks.agentTurnEnd = append(p.Hooks.agentTurnEnd, h)
	p.Hooks.mark("agent_turn_end")
}

// OnError registers a runtime-error hook.
func (p *Pipeline) OnError(h func(payload map[string]any)) {
	p.Hooks.errorHooks = append(p.Hooks.errorHooks, h)
	p.Hooks.mark("error")
}

// OnRecordingStarted registers a recording-started hook.
func (p *Pipeline) OnRecordingStarted(h func(status map[string]any)) {
	p.Hooks.recordingStarted = append(p.Hooks.recordingStarted, h)
	p.Hooks.mark("recording_started")
}

// OnRecordingStopped registers a recording-stopped hook.
func (p *Pipeline) OnRecordingStopped(h func(status map[string]any)) {
	p.Hooks.recordingStopped = append(p.Hooks.recordingStopped, h)
	p.Hooks.mark("recording_stopped")
}

// OnRecordingFailed registers a recording-failed hook.
func (p *Pipeline) OnRecordingFailed(h func(status map[string]any)) {
	p.Hooks.recordingFailed = append(p.Hooks.recordingFailed, h)
	p.Hooks.mark("recording_failed")
}

// OnVisionFrame registers a per-participant video frame hook.
func (p *Pipeline) OnVisionFrame(h func(frame map[string]any)) {
	p.Hooks.visionFrame = append(p.Hooks.visionFrame, h)
	p.Hooks.mark("vision_frame")
}

// OnAudioFrame registers a raw audio-delta hook.
func (p *Pipeline) OnAudioFrame(h func(frame map[string]any)) {
	p.Hooks.audioDelta = append(p.Hooks.audioDelta, h)
	p.Hooks.mark("audio_delta")
}

// PushVisionFrame buffers a vision frame and fires vision-frame hooks.
func (p *Pipeline) PushVisionFrame(frame map[string]any) {
	p.frameMu.Lock()
	p.frameBuffer = append(p.frameBuffer, frame)
	if len(p.frameBuffer) > visionFrameBufferMax {
		p.frameBuffer = p.frameBuffer[len(p.frameBuffer)-visionFrameBufferMax:]
	}
	p.frameMu.Unlock()
	for _, h := range p.Hooks.visionFrame {
		fireFrameHook(h, frame)
	}
}

// PushAudioFrame fires audio-delta hooks.
func (p *Pipeline) PushAudioFrame(frame map[string]any) {
	for _, h := range p.Hooks.audioDelta {
		fireFrameHook(h, frame)
	}
}

// GetLatestFrames returns up to n most-recent buffered vision frames.
func (p *Pipeline) GetLatestFrames(n int) []map[string]any {
	if n <= 0 {
		return nil
	}
	p.frameMu.Lock()
	defer p.frameMu.Unlock()
	if n > len(p.frameBuffer) {
		n = len(p.frameBuffer)
	}
	out := make([]map[string]any, n)
	copy(out, p.frameBuffer[len(p.frameBuffer)-n:])
	return out
}

func fireFrameHook(h func(map[string]any), frame map[string]any) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("frame hook panicked: %v", r)
		}
	}()
	h(frame)
}

// Config returns the resolved pipeline topology (mode + active components).
func (p *Pipeline) Config() PipelineConfigInfo {
	components := map[PipelineComponent]bool{}
	if p.STT != nil {
		components[ComponentSTT] = true
	}
	if p.LLM != nil {
		components[ComponentLLM] = true
	}
	if p.TTS != nil {
		components[ComponentTTS] = true
	}
	if p.VAD != nil {
		components[ComponentVAD] = true
	}
	if p.TurnDetector != nil {
		components[ComponentTurnDetector] = true
	}
	if p.Denoise != nil {
		components[ComponentDenoise] = true
	}
	isRealtime := llmIsRealtime(p.LLM) || (p.RealtimeConfig != nil && p.RealtimeConfig.Mode != "")
	hasSTT, hasLLM, hasTTS := p.STT != nil, p.LLM != nil, p.TTS != nil
	if isRealtime {
		components[ComponentRealtimeModel] = true
		rtMode := RealtimeModeFullS2S
		if p.RealtimeConfig != nil {
			switch p.RealtimeConfig.Mode {
			case "hybrid_stt":
				rtMode = RealtimeModeHybridSTT
			case "hybrid_tts":
				rtMode = RealtimeModeHybridTTS
			case "llm_only":
				rtMode = RealtimeModeLLMOnly
			}
		}
		mode := PipelineModeRealtime
		if hasSTT || hasTTS {
			mode = PipelineModeHybrid
		}
		return PipelineConfigInfo{Mode: mode, RealtimeMode: rtMode, IsRealtime: true, ActiveComponents: components}
	}
	var mode PipelineMode
	switch {
	case hasSTT && hasLLM && hasTTS:
		mode = PipelineModeFullCascading
	case hasSTT && hasLLM:
		mode = PipelineModeSTTLLMOnly
	case hasLLM && hasTTS:
		mode = PipelineModeLLMTTSOnly
	case hasSTT && hasTTS:
		mode = PipelineModeSTTTTSOnly
	case hasSTT:
		mode = PipelineModeSTTOnly
	case hasLLM:
		mode = PipelineModeLLMOnly
	case hasTTS:
		mode = PipelineModeTTSOnly
	default:
		mode = PipelineModePartialCascading
	}
	return PipelineConfigInfo{Mode: mode, IsRealtime: false, ActiveComponents: components}
}

// PipelineConfigInfo describes the resolved pipeline topology.
type PipelineConfigInfo struct {
	Mode             PipelineMode
	RealtimeMode     RealtimeMode
	IsRealtime       bool
	ActiveComponents map[PipelineComponent]bool
}

// HasComponent reports whether a component is active.
func (c PipelineConfigInfo) HasComponent(comp PipelineComponent) bool {
	return c.ActiveComponents[comp]
}

// ChangePipeline mutates pipeline slots at runtime (nil values are ignored).
func (p *Pipeline) ChangePipeline(opts PipelineOptions) {
	if opts.STT != nil {
		p.STT = opts.STT
	}
	if opts.LLM != nil {
		p.LLM = opts.LLM
	}
	if opts.TTS != nil {
		p.TTS = opts.TTS
	}
	if opts.VAD != nil {
		p.VAD = opts.VAD
	}
	if opts.TurnDetector != nil {
		p.TurnDetector = opts.TurnDetector
	}
	if opts.Denoise != nil {
		p.Denoise = opts.Denoise
	}
}

// SendMessage speaks text through the bound session.
func (p *Pipeline) SendMessage(ctx context.Context, text string) error {
	if p.agent != nil && p.agent.base().session != nil {
		_, err := p.agent.base().session.Say(ctx, text)
		return err
	}
	return nil
}

// ReplyWithContext triggers a context-aware reply.
func (p *Pipeline) ReplyWithContext(ctx context.Context, instructions string) error {
	if p.agent != nil && p.agent.base().session != nil {
		_, err := p.agent.base().session.Reply(ctx, instructions, true)
		return err
	}
	return nil
}

// ProcessText generates a response to text.
func (p *Pipeline) ProcessText(ctx context.Context, text string) error {
	if p.agent != nil && p.agent.base().session != nil {
		return p.agent.base().session.Generate(ctx, text)
	}
	return nil
}

// Interrupt interrupts the bound session.
func (p *Pipeline) Interrupt() {
	if p.agent != nil && p.agent.base().session != nil {
		p.agent.base().session.Interrupt(false)
	}
}
