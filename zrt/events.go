package zrt

import (
	"cmp"
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
)

var recordingStateNames = map[pb.RecordingState]string{
	0: "idle", 1: "recording", 2: "finalizing", 3: "uploading", 4: "completed", 5: "failed",
}

func recordingStateToString(state pb.RecordingState) string {
	if v, ok := recordingStateNames[state]; ok {
		return v
	}
	return "unknown_" + strconv.Itoa(int(state))
}

var unrecoverableAuthPatterns = []string{
	"invalid credentials", "invalid api key", "api_key_invalid", "unauthorized", "401", "403",
	"permission_denied", "permission denied", "insufficient permission", "forbidden",
	"authentication failed", "auth failed", "auth error",
}

func isUnrecoverableAuthError(component, message string) bool {
	if message == "" {
		return false
	}
	if component != "llm" && component != "stt" && component != "tts" {
		return false
	}
	lower := strings.ToLower(message)
	for _, p := range unrecoverableAuthPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// ---- shared event handlers (identical between direct and registered modes) ----

func evGenerationStarted(s *AgentSession, gs *pb.GenerationStartedEvent) {
	for _, h := range s.pipeline.Hooks.generationStarted {
		safeHook("generation_started", func() { h(gs.GetTurnNumber()) })
	}
	s.Emit("generation_started", map[string]any{"turn_number": gs.GetTurnNumber()})
}

func evGenerationComplete(s *AgentSession, gc *pb.GenerationCompleteEvent) {
	for _, h := range s.pipeline.Hooks.generationComplete {
		safeHook("generation_complete", func() { h(gc.GetTurnNumber(), gc.GetWasInterrupted()) })
	}
	s.Emit("generation_complete", map[string]any{"turn_number": gc.GetTurnNumber(), "was_interrupted": gc.GetWasInterrupted()})
}

func evGenerationChunk(s *AgentSession, gc *pb.GenerationChunkEvent) {
	meta := copyAnyMap(gc.GetMetadata())
	for _, h := range s.pipeline.Hooks.generationChunk {
		safeHook("generation_chunk", func() { h(gc.GetText(), meta) })
	}
	s.Emit("generation_chunk", map[string]any{"text": gc.GetText(), "metadata": meta})
}

func evSynthesisStarted(s *AgentSession, ss *pb.SynthesisStartedEvent) {
	for _, h := range s.pipeline.Hooks.synthesisStarted {
		safeHook("synthesis_started", func() { h(ss.GetText()) })
	}
	s.Emit("synthesis_started", map[string]any{"text": ss.GetText()})
}

func evSynthesisInterrupted(s *AgentSession, si *pb.SynthesisInterruptedEvent) {
	for _, h := range s.pipeline.Hooks.synthesisInterrupted {
		safeHook("synthesis_interrupted", func() { h(si.GetReason()) })
	}
	s.Emit("synthesis_interrupted", map[string]any{"reason": si.GetReason()})
}

func evFirstAudioByte(s *AgentSession, fab *pb.FirstAudioByteEvent) {
	if obs, ok := s.pipeline.TTS.(firstAudioObservable); ok {
		if cb := obs.firstAudioByteCallback(); cb != nil {
			cb(fab.GetTtfbMs(), fab.GetByteCount())
		}
	}
	for _, h := range s.pipeline.Hooks.firstAudioByte {
		safeHook("first_audio_byte", func() { h(fab.GetTtfbMs(), fab.GetByteCount()) })
	}
	s.Emit("first_audio_byte", map[string]any{"ttfb_ms": fab.GetTtfbMs(), "byte_count": fab.GetByteCount()})
}

func evLastAudioByte(s *AgentSession, lab *pb.LastAudioByteEvent) {
	for _, h := range s.pipeline.Hooks.lastAudioByte {
		safeHook("last_audio_byte", func() { h(float64(lab.GetDurationSeconds())) })
	}
	s.Emit("last_audio_byte", map[string]any{"duration_seconds": lab.GetDurationSeconds()})
}

func evWordTiming(s *AgentSession, wt *pb.WordTimingEvent) {
	for _, h := range s.pipeline.Hooks.wordTiming {
		safeHook("word_timing", func() {
			h(wt.GetWord(), float64(wt.GetStartSeconds()), float64(wt.GetEndSeconds()), wt.GetCumulativeText())
		})
	}
	s.Emit("word_timing", map[string]any{"word": wt.GetWord(), "start_seconds": wt.GetStartSeconds(), "end_seconds": wt.GetEndSeconds(), "cumulative_text": wt.GetCumulativeText()})
}

func evTTSCapabilities(s *AgentSession, tc *pb.TtsCapabilitiesEvent) {
	caps := map[string]any{"can_pause": tc.GetCanPause(), "supports_word_timestamps": tc.GetSupportsWordTimestamps(), "sample_rate": tc.GetSampleRate(), "num_channels": tc.GetNumChannels()}
	s.mu.Lock()
	s.ttsCapabilities = caps
	s.mu.Unlock()
	s.Emit("tts_capabilities", caps)
}

func evTranscriptPreflight(s *AgentSession, tp *pb.TranscriptPreflightEvent) {
	for _, h := range s.pipeline.Hooks.transcriptPreflight {
		safeHook("transcript_preflight", func() { h(tp.GetText()) })
	}
	s.Emit("transcript_preflight", map[string]any{"text": tp.GetText()})
}

func evEOUDetected(s *AgentSession, ed *pb.EouDetectedEvent) {
	s.mu.Lock()
	s.lastEOU = map[string]any{"probability": ed.GetProbability(), "wait_ms": ed.GetWaitMs(), "text": ed.GetText()}
	s.mu.Unlock()
	for _, h := range s.pipeline.Hooks.eouDetected {
		safeHook("eou_detected", func() { h(float64(ed.GetProbability()), ed.GetWaitMs(), ed.GetText()) })
	}
	s.Emit("eou_detected", map[string]any{"probability": ed.GetProbability(), "wait_ms": ed.GetWaitMs(), "text": ed.GetText()})
}

func evSTTStreamStarted(s *AgentSession, ss *pb.SttStreamStartedEvent) {
	s.Emit("stt_stream_started", map[string]any{"provider": ss.GetProvider(), "model": ss.GetModel()})
}

func evSTTStreamEnded(s *AgentSession, se *pb.SttStreamEndedEvent) {
	s.Emit("stt_stream_ended", map[string]any{"provider": se.GetProvider(), "reason": se.GetReason()})
}

func evVADEvent(s *AgentSession, ve *pb.VadEventProto) {
	isSpeech := strings.ToLower(ve.GetKind()) == "speech_start"
	et := VADEventEndOfSpeech
	if isSpeech {
		et = VADEventStartOfSpeech
	}
	ts := float64(ve.GetTimestampUs()) / 1_000_000.0
	data := VADData{IsSpeech: isSpeech, Confidence: float64(ve.GetConfidence()), Timestamp: ts, SpeechDuration: float64(ve.GetSpeechDurationS()), SilenceDuration: float64(ve.GetSilenceDurationS())}
	if obs, ok := s.pipeline.VAD.(vadObservable); ok {
		if cb := obs.vadCallback(); cb != nil {
			cb(VADResponse{EventType: et, Data: data})
		}
	}
	s.Emit("vad_event", map[string]any{"event_type": string(et), "is_speech": isSpeech, "timestamp": ts, "confidence": data.Confidence, "speech_duration": data.SpeechDuration, "silence_duration": data.SilenceDuration})
}

func evInterrupt(s *AgentSession, i *pb.InterruptEvent) {
	if u := s.CurrentUtterance(); u != nil {
		u.Interrupt(false)
	}
	s.Emit("interrupt", map[string]any{"reason": i.GetReason(), "partial_response": i.GetPartialResponse()})
}

func evStateChange(s *AgentSession, sc *pb.SessionStateChange) {
	switch strings.ToLower(sc.GetState()) {
	case "idle":
		s.updateAgentState(AgentStateIdle)
	case "listening":
		s.updateAgentState(AgentStateListening)
	case "generating":
		s.updateAgentState(AgentStateThinking)
	case "speaking":
		s.updateAgentState(AgentStateSpeaking)
	}
}

func evSayComplete(s *AgentSession) {
	if u := s.CurrentUtterance(); u != nil {
		u.markDone()
	}
}

func evRecordingStatus(s *AgentSession, rs *pb.RecordingStatusEvent) {
	status := map[string]any{
		"recording_id":     rs.GetRecordingId(),
		"state":            recordingStateToString(rs.GetState()),
		"output_uri":       rs.GetOutputUri(),
		"duration_seconds": rs.GetDurationSeconds(),
		"file_size_bytes":  rs.GetFileSizeBytes(),
		"error_message":    rs.GetErrorMessage(),
		"transcript_uri":   rs.GetTranscriptUri(),
		"metadata":         copyAnyMap(rs.GetMetadata()),
		"track_kind":       rs.GetTrackKind(),
	}
	logger.Infof("Recording status: %s (id=%s)", status["state"], rs.GetRecordingId())
	dispatchRecordingStatus(s, status)
}

func dispatchRecordingStatus(s *AgentSession, status map[string]any) {
	s.updateRecordingState(status)
	state, _ := status["state"].(string)
	var hookList []func(map[string]any)
	switch state {
	case "recording":
		hookList = s.pipeline.Hooks.recordingStarted
		s.pipeline.Emit("recording_started", status)
	case "completed":
		hookList = s.pipeline.Hooks.recordingStopped
		s.pipeline.Emit("recording_stopped", status)
	case "failed":
		hookList = s.pipeline.Hooks.recordingFailed
		s.pipeline.Emit("recording_failed", status)
	}
	for _, h := range hookList {
		safeHook("recording_status", func() { h(status) })
	}
}

func evWarning(s *AgentSession, w *pb.WarningEvent) {
	logger.Warnf("[runtime warning] %s: %s", w.GetCode(), w.GetMessage())
	s.Emit("runtime_warning", map[string]any{"code": w.GetCode(), "message": w.GetMessage()})
}

func evMetrics(s *AgentSession, m *pb.MetricsSnapshot) {
	nanToNil := func(x float32) any {
		if math.IsNaN(float64(x)) {
			return nil
		}
		return x
	}
	var sttTTFB any
	if m.SttTtfbMs != nil {
		sttTTFB = nanToNil(m.GetSttTtfbMs())
	}
	payload := map[string]any{
		"turn_number":           m.GetTurnNumber(),
		"stt_ttfb_ms":           sttTTFB,
		"llm_ttfb_ms":           nanToNil(m.GetLlmTtfbMs()),
		"tts_ttfb_ms":           nanToNil(m.GetTtsTtfbMs()),
		"total_turn_latency_ms": nanToNil(m.GetTotalTurnLatencyMs()),
		"tokens_in":             m.GetTokensIn(),
		"tokens_out":            m.GetTokensOut(),
		"cached_tokens":         m.GetCachedTokens(),
		"tokens_total":          m.GetTokensTotal(),
		"input_text_tokens":     m.GetInputTextTokens(),
		"input_audio_tokens":    m.GetInputAudioTokens(),
		"input_image_tokens":    m.GetInputImageTokens(),
		"output_text_tokens":    m.GetOutputTextTokens(),
		"output_audio_tokens":   m.GetOutputAudioTokens(),
		"output_image_tokens":   m.GetOutputImageTokens(),
		"cached_text_tokens":    m.GetCachedTextTokens(),
		"cached_audio_tokens":   m.GetCachedAudioTokens(),
		"cached_image_tokens":   m.GetCachedImageTokens(),
		"thoughts_tokens":       m.GetThoughtsTokens(),
	}
	metricsCollector.append(payload)
	s.Emit("metrics_collected", payload)
}

func evDTMF(s *AgentSession, d *pb.DTMFEvent) {
	s.Emit("dtmf_received", map[string]any{"digit": d.GetDigit(), "participant_id": d.GetParticipantId()})
}

func evVisionFrame(s *AgentSession, v *pb.VisionFrameEvent) {
	frame := map[string]any{"data": v.GetData(), "mime_type": v.GetMimeType(), "width": v.GetWidth(), "height": v.GetHeight(), "participant_id": v.GetParticipantId(), "timestamp_ms": v.GetTimestampMs()}
	s.pipeline.PushVisionFrame(frame)
}

func evAudioFrame(s *AgentSession, a *pb.AudioFrameEvent) {
	frame := map[string]any{"pcm": a.GetPcm(), "sample_rate": a.GetSampleRate(), "source": a.GetSource(), "participant_id": a.GetParticipantId(), "timestamp_ms": a.GetTimestampMs()}
	s.pipeline.PushAudioFrame(frame)
}

func evStreamEvent(s *AgentSession, se *pb.StreamEventProto) {
	payload := map[string]any{"participant_id": se.GetParticipantId(), "kind": se.GetKind(), "enabled": se.GetEnabled()}
	if se.GetEnabled() {
		s.Emit("stream_enabled", payload)
	} else {
		s.Emit("stream_disabled", payload)
	}
}

func evSignaling(s *AgentSession, sid string) {
	if sid == "" {
		return
	}
	s.setSignalingSessionID(sid)
	logger.Infof("Signaling session_id assigned: %s", sid)
	s.Emit("signaling_session_assigned", map[string]any{"session_id": sid})
}

func evVoicemail(s *AgentSession, v *pb.VoicemailDetectedEvent) {
	s.Emit("voicemail_detected", map[string]any{"confidence": v.GetConfidence(), "transcript": v.GetTranscript()})
}

func evPubSubMessage(s *AgentSession, m *pb.PubSubMessageEvent) {
	var payload any
	if pj := m.GetPayloadJson(); pj != "" {
		if err := json.Unmarshal([]byte(pj), &payload); err != nil {
			payload = pj
		}
	}
	s.Emit("pubsub_message", map[string]any{
		"topic":       m.GetTopic(),
		"message":     m.GetMessage(),
		"sender_id":   m.GetSenderId(),
		"sender_name": m.GetSenderName(),
		"timestamp":   m.GetTimestamp(),
		"payload":     payload,
	})
}

func evAgentSwitched(s *AgentSession, sw *pb.AgentSwitchedEvent) {
	s.applyAgentSwitch(sw.GetFrom(), sw.GetTo(), sw.GetReason())
	s.Emit("agent_switched", map[string]any{"from": sw.GetFrom(), "to": sw.GetTo(), "reason": sw.GetReason()})
}

func evKBHits(s *AgentSession, k *pb.KbHitsEvent) {
	var docs []map[string]any
	for _, d := range k.GetDocuments() {
		docs = append(docs, map[string]any{"id": d.GetId(), "content": d.GetContent(), "score": d.GetScore(), "metadata": copyAnyMap(d.GetMetadata())})
	}
	s.Emit("kb_hits", map[string]any{"query": k.GetQuery(), "documents": docs, "latency_ms": k.GetLatencyMs()})
}

func evParticipant(s *AgentSession, p *pb.ParticipantEventProto) {
	payload := map[string]any{"type": p.GetType(), "participant_id": p.GetParticipantId(), "display_name": p.GetDisplayName()}
	t := strings.ToLower(p.GetType())
	if strings.Contains(t, "join") {
		s.Emit("participant_joined", payload)
	} else if strings.Contains(t, "left") || strings.Contains(t, "leave") {
		s.Emit("participant_left", payload)
	}
}

var agentStateMap = map[string]AgentState{
	"starting": AgentStateIdle, "idle": AgentStateIdle, "listening": AgentStateListening,
	"thinking": AgentStateThinking, "speaking": AgentStateSpeaking, "closing": AgentStateIdle,
}

var userStateMap = map[string]UserState{
	"idle": UserStateIdle, "speaking": UserStateSpeaking, "listening": UserStateListening,
}

func evAgentStateChanged(s *AgentSession, sc *pb.AgentStateChangedEvent) {
	st := strings.ToLower(sc.GetState())
	if v, ok := agentStateMap[st]; ok {
		s.updateAgentState(v)
	}
	s.Emit("agent_state_changed", map[string]any{"state": st, "reason": sc.GetReason()})
}

func evUserStateChanged(s *AgentSession, uc *pb.UserStateChangedEvent) {
	st := strings.ToLower(uc.GetState())
	if v, ok := userStateMap[st]; ok {
		s.updateUserState(v)
	}
	s.Emit("user_state_changed", map[string]any{"state": st, "reason": uc.GetReason()})
}

func evLLMCompleted(s *AgentSession, lc *pb.LLMCompletedEvent) {
	for _, h := range s.pipeline.Hooks.llmCompleted {
		safeHook("llm_completed", func() { h(lc.GetResponseText(), lc.GetInterrupted()) })
	}
	s.Emit("llm_completed", map[string]any{"response_text": lc.GetResponseText(), "interrupted": lc.GetInterrupted()})
}

// ---- transcript / agent_speech / turn_complete (mode-divergent) ----

func buildSTTResponse(t *pb.TranscriptEvent) STTResponse {
	startS := float64(t.GetStartTimeUs()) / 1_000_000.0
	endS := float64(t.GetEndTimeUs()) / 1_000_000.0
	duration := 0.0
	if endS > 0 && startS > 0 && endS >= startS {
		duration = endS - startS
	}
	lang := t.GetLanguage()
	return STTResponse{
		EventType: pick(t.GetIsFinal(), SpeechEventFinal, SpeechEventInterim),
		Data:      SpeechData{Text: t.GetText(), Confidence: float64(t.GetConfidence()), Language: lang, StartTime: startS, EndTime: endS, Duration: duration},
		Metadata: map[string]string{
			"participant_id": t.GetParticipantId(),
			"turn_resumed":   strconv.FormatBool(t.GetTurnResumed()),
			"item_id":        t.GetItemId(),
		},
	}
}

func transcriptDirect(s *AgentSession, t *pb.TranscriptEvent) {
	resp := buildSTTResponse(t)
	if obs, ok := s.pipeline.STT.(transcriptObservable); ok {
		if cb := obs.transcriptCallback(); cb != nil {
			cb(resp)
		}
	}
	s.pushSTTObservation(resp)
	if t.GetIsFinal() {
		s.updateUserState(UserStateSpeaking)
	}
}

func transcriptRegistered(s *AgentSession, t *pb.TranscriptEvent) {
	resp := buildSTTResponse(t)
	s.pushSTTObservation(resp)
	if t.GetIsFinal() {
		for _, h := range s.pipeline.Hooks.userTurnStart {
			safeHook("user_turn_start", func() { h(t.GetText()) })
		}
		s.Emit("user_turn_start", map[string]any{"text": t.GetText()})
		s.updateUserState(UserStateSpeaking)
	}
}

func agentSpeechDirect(s *AgentSession, a *pb.AgentSpeechEvent) {
	role := cmp.Or(a.GetRole(), "assistant")
	if a.GetIsFinal() {
		s.updateAgentState(AgentStateIdle)
		s.mu.Lock()
		s.lastAgentSpeech = map[string]any{"text": a.GetText(), "role": role, "item_id": a.GetItemId()}
		s.mu.Unlock()
	} else {
		s.updateAgentState(AgentStateSpeaking)
	}
}

func agentSpeechRegistered(s *AgentSession, a *pb.AgentSpeechEvent) {
	if a.GetIsFinal() {
		for _, h := range s.pipeline.Hooks.agentTurnEnd {
			safeHook("agent_turn_end", func() { h() })
		}
		s.Emit("agent_turn_end", map[string]any{"text": a.GetText()})
		s.updateAgentState(AgentStateIdle)
	} else {
		if s.AgentState() != AgentStateSpeaking {
			s.Emit("agent_turn_start", map[string]any{"partial": a.GetText()})
		}
		s.updateAgentState(AgentStateSpeaking)
	}
}

func turnCompleteDirect(s *AgentSession, tc *pb.TurnComplete) {
	s.Emit("turn_complete", turnCompletePayload(tc))
	if u := s.CurrentUtterance(); u != nil {
		u.markDone()
	}
}

func turnCompleteRegistered(s *AgentSession, tc *pb.TurnComplete) {
	for _, h := range s.pipeline.Hooks.userTurnEnd {
		safeHook("user_turn_end", func() { h() })
	}
	s.Emit("user_turn_end", turnCompletePayload(tc))
	if u := s.CurrentUtterance(); u != nil {
		u.markDone()
	}
}

func turnCompletePayload(tc *pb.TurnComplete) map[string]any {
	return map[string]any{
		"user_transcript":  tc.GetUserTranscript(),
		"agent_transcript": tc.GetAgentTranscript(),
		"tool_calls_count": tc.GetToolCallsCount(),
		"total_latency_ms": tc.GetTotalLatencyMs(),
		"context_messages": tc.GetContextMessages(),
		"context_tokens":   tc.GetContextTokens(),
	}
}

// ---- bridge-only synthesized turn events ----

func evUserTurnStart(s *AgentSession, uts *pb.UserTurnStartEvent) {
	for _, h := range s.pipeline.Hooks.userTurnStart {
		safeHook("user_turn_start", func() { h(uts.GetTranscript()) })
	}
	s.Emit("user_turn_start", map[string]any{"text": uts.GetTranscript()})
}

func evUserTurnEnd(s *AgentSession, ute *pb.UserTurnEndEvent) {
	for _, h := range s.pipeline.Hooks.userTurnEnd {
		safeHook("user_turn_end", func() { h() })
	}
	s.Emit("user_turn_end", map[string]any{"response_text": ute.GetResponseText(), "interrupted": ute.GetInterrupted()})
}

func evAgentTurnStart(s *AgentSession, ats *pb.AgentTurnStartEvent) {
	for _, h := range s.pipeline.Hooks.agentTurnStart {
		safeHook("agent_turn_start", func() { h() })
	}
	s.Emit("agent_turn_start", map[string]any{"turn_number": ats.GetTurnNumber()})
}

func evAgentTurnEnd(s *AgentSession, ate *pb.AgentTurnEndEvent) {
	for _, h := range s.pipeline.Hooks.agentTurnEnd {
		safeHook("agent_turn_end", func() { h() })
	}
	s.mu.Lock()
	cached := s.lastAgentSpeech
	s.mu.Unlock()
	text, role, itemID := "", "assistant", ""
	if cached != nil {
		text, _ = cached["text"].(string)
		if r, ok := cached["role"].(string); ok {
			role = r
		}
		itemID, _ = cached["item_id"].(string)
	}
	s.Emit("agent_turn_end", map[string]any{"turn_number": ate.GetTurnNumber(), "interrupted": ate.GetInterrupted(), "text": text, "role": role, "item_id": itemID})
}

// ---- tool execution (shared) ----

func toolFiller(tools []*FunctionTool, toolName string) string {
	for _, t := range tools {
		if t != nil && t.Info.Name == toolName {
			return t.Info.Filler
		}
	}
	return ""
}

const toolFillerGraceMs = 300

// toolFillerGracePeriod returns the per-tool grace period (ms) before the filler is
// spoken, or 0 when the tool has no override (callers fall back to toolFillerGraceMs).
func toolFillerGracePeriod(tools []*FunctionTool, toolName string) int {
	for _, t := range tools {
		if t != nil && t.Info.Name == toolName {
			return t.Info.FillerGracePeriod
		}
	}
	return 0
}

func executeToolWithFiller(ctx context.Context, tools []*FunctionTool, callID, toolName, argsJSON, filler string, graceMs int, say func(text string), sendResult func(callID, resultJSON string, isErr bool), onResult func(any) any) {
	done := make(chan struct{})
	if filler != "" && say != nil {
		grace := graceMs
		if grace <= 0 {
			grace = toolFillerGraceMs
		}
		go func() {
			timer := time.NewTimer(time.Duration(grace) * time.Millisecond)
			defer timer.Stop()
			select {
			case <-timer.C:
				defer func() {
					if r := recover(); r != nil {
						logger.Debugf("tool filler say panicked: %v", r)
					}
				}()
				say(filler)
			case <-done:
			}
		}()
	}
	executeTool(ctx, tools, callID, toolName, argsJSON, sendResult, onResult)
	close(done)
}

// executeTool runs the named tool and sends its result. onResult, when non-nil,
// is applied to the raw return value before serialization; it is how a tool that
// returns an Agent object is converted into an agent-switch marker.
func executeTool(ctx context.Context, tools []*FunctionTool, callID, toolName, argsJSON string, sendResult func(callID, resultJSON string, isErr bool), onResult func(any) any) {
	var args map[string]any
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			args = map[string]any{}
		}
	}
	if args == nil {
		args = map[string]any{}
	}
	var tool *FunctionTool
	for _, t := range tools {
		if t != nil && t.Info.Name == toolName {
			tool = t
			break
		}
	}
	if tool == nil || tool.Handler == nil {
		b, _ := json.Marshal(map[string]any{"error": "Tool '" + toolName + "' not found"})
		sendResult(callID, string(b), true)
		return
	}
	result, err := func() (res any, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = errFromRecover(r)
			}
		}()
		return tool.Handler(ctx, args)
	}()
	if err != nil {
		logger.Errorf("Tool '%s' error: %v", toolName, err)
		b, _ := json.Marshal(map[string]any{"error": err.Error()})
		sendResult(callID, string(b), true)
		return
	}
	if onResult != nil {
		result = onResult(result)
	}
	var resultJSON string
	if str, ok := result.(string); ok {
		resultJSON = str
	} else {
		b, mErr := json.Marshal(result)
		if mErr != nil {
			logger.Errorf("Tool '%s' result not serializable: %v", toolName, mErr)
			eb, _ := json.Marshal(map[string]any{"error": mErr.Error()})
			sendResult(callID, string(eb), true)
			return
		}
		resultJSON = string(b)
	}
	sendResult(callID, resultJSON, false)
}

func errFromRecover(r any) error {
	if e, ok := r.(error); ok {
		return e
	}
	return &recoverError{v: r}
}

type recoverError struct{ v any }

func (e *recoverError) Error() string {
	b, _ := json.Marshal(map[string]any{"panic": strconv.Quote(strings.TrimSpace(stringify(e.v)))})
	return string(b)
}

func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case error:
		return x.Error()
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func pick[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func copyAnyMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// pushSTTObservation forwards a transcript to a registered custom STT observation
// channel, if any.
func (s *AgentSession) pushSTTObservation(resp STTResponse) {
	s.mu.Lock()
	ch := s.sttObservationQueue
	s.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- resp:
	default:
		logger.Debugf("stt observation queue full — dropping transcript (is_final=%v)", resp.EventType == SpeechEventFinal)
	}
}
