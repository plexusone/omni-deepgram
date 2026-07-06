package realtime

import (
	agentinterfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/agent/v1/websocket/interfaces"
)

// chanHandler implements agentinterfaces.AgentMessageChan for receiving events from Deepgram.
type chanHandler struct {
	binaryChan               chan *[]byte
	openChan                 chan *agentinterfaces.OpenResponse
	welcomeChan              chan *agentinterfaces.WelcomeResponse
	conversationTextChan     chan *agentinterfaces.ConversationTextResponse
	userStartedSpeakingChan  chan *agentinterfaces.UserStartedSpeakingResponse
	agentThinkingChan        chan *agentinterfaces.AgentThinkingResponse
	functionCallRequestChan  chan *agentinterfaces.FunctionCallRequestResponse
	agentStartedSpeakingChan chan *agentinterfaces.AgentStartedSpeakingResponse
	agentAudioDoneChan       chan *agentinterfaces.AgentAudioDoneResponse
	injectionRefusedChan     chan *agentinterfaces.InjectionRefusedResponse
	keepAliveChan            chan *agentinterfaces.KeepAlive
	settingsAppliedChan      chan *agentinterfaces.SettingsAppliedResponse
	closeChan                chan *agentinterfaces.CloseResponse
	errorChan                chan *agentinterfaces.ErrorResponse
	unhandledChan            chan *[]byte
}

// newChanHandler creates a new chanHandler with buffered channels.
func newChanHandler() *chanHandler {
	const bufferSize = 100
	return &chanHandler{
		binaryChan:               make(chan *[]byte, bufferSize),
		openChan:                 make(chan *agentinterfaces.OpenResponse, bufferSize),
		welcomeChan:              make(chan *agentinterfaces.WelcomeResponse, bufferSize),
		conversationTextChan:     make(chan *agentinterfaces.ConversationTextResponse, bufferSize),
		userStartedSpeakingChan:  make(chan *agentinterfaces.UserStartedSpeakingResponse, bufferSize),
		agentThinkingChan:        make(chan *agentinterfaces.AgentThinkingResponse, bufferSize),
		functionCallRequestChan:  make(chan *agentinterfaces.FunctionCallRequestResponse, bufferSize),
		agentStartedSpeakingChan: make(chan *agentinterfaces.AgentStartedSpeakingResponse, bufferSize),
		agentAudioDoneChan:       make(chan *agentinterfaces.AgentAudioDoneResponse, bufferSize),
		injectionRefusedChan:     make(chan *agentinterfaces.InjectionRefusedResponse, bufferSize),
		keepAliveChan:            make(chan *agentinterfaces.KeepAlive, bufferSize),
		settingsAppliedChan:      make(chan *agentinterfaces.SettingsAppliedResponse, bufferSize),
		closeChan:                make(chan *agentinterfaces.CloseResponse, bufferSize),
		errorChan:                make(chan *agentinterfaces.ErrorResponse, bufferSize),
		unhandledChan:            make(chan *[]byte, bufferSize),
	}
}

// GetBinary returns the binary (audio) channels.
func (h *chanHandler) GetBinary() []*chan *[]byte {
	return []*chan *[]byte{&h.binaryChan}
}

// GetOpen returns the open response channels.
func (h *chanHandler) GetOpen() []*chan *agentinterfaces.OpenResponse {
	return []*chan *agentinterfaces.OpenResponse{&h.openChan}
}

// GetWelcome returns the welcome response channels.
func (h *chanHandler) GetWelcome() []*chan *agentinterfaces.WelcomeResponse {
	return []*chan *agentinterfaces.WelcomeResponse{&h.welcomeChan}
}

// GetConversationText returns the conversation text response channels.
func (h *chanHandler) GetConversationText() []*chan *agentinterfaces.ConversationTextResponse {
	return []*chan *agentinterfaces.ConversationTextResponse{&h.conversationTextChan}
}

// GetUserStartedSpeaking returns the user started speaking response channels.
func (h *chanHandler) GetUserStartedSpeaking() []*chan *agentinterfaces.UserStartedSpeakingResponse {
	return []*chan *agentinterfaces.UserStartedSpeakingResponse{&h.userStartedSpeakingChan}
}

// GetAgentThinking returns the agent thinking response channels.
func (h *chanHandler) GetAgentThinking() []*chan *agentinterfaces.AgentThinkingResponse {
	return []*chan *agentinterfaces.AgentThinkingResponse{&h.agentThinkingChan}
}

// GetFunctionCallRequest returns the function call request response channels.
func (h *chanHandler) GetFunctionCallRequest() []*chan *agentinterfaces.FunctionCallRequestResponse {
	return []*chan *agentinterfaces.FunctionCallRequestResponse{&h.functionCallRequestChan}
}

// GetAgentStartedSpeaking returns the agent started speaking response channels.
func (h *chanHandler) GetAgentStartedSpeaking() []*chan *agentinterfaces.AgentStartedSpeakingResponse {
	return []*chan *agentinterfaces.AgentStartedSpeakingResponse{&h.agentStartedSpeakingChan}
}

// GetAgentAudioDone returns the agent audio done response channels.
func (h *chanHandler) GetAgentAudioDone() []*chan *agentinterfaces.AgentAudioDoneResponse {
	return []*chan *agentinterfaces.AgentAudioDoneResponse{&h.agentAudioDoneChan}
}

// GetInjectionRefused returns the injection refused response channels.
func (h *chanHandler) GetInjectionRefused() []*chan *agentinterfaces.InjectionRefusedResponse {
	return []*chan *agentinterfaces.InjectionRefusedResponse{&h.injectionRefusedChan}
}

// GetKeepAlive returns the keep alive response channels.
func (h *chanHandler) GetKeepAlive() []*chan *agentinterfaces.KeepAlive {
	return []*chan *agentinterfaces.KeepAlive{&h.keepAliveChan}
}

// GetSettingsApplied returns the settings applied response channels.
func (h *chanHandler) GetSettingsApplied() []*chan *agentinterfaces.SettingsAppliedResponse {
	return []*chan *agentinterfaces.SettingsAppliedResponse{&h.settingsAppliedChan}
}

// GetClose returns the close response channels.
func (h *chanHandler) GetClose() []*chan *agentinterfaces.CloseResponse {
	return []*chan *agentinterfaces.CloseResponse{&h.closeChan}
}

// GetError returns the error response channels.
func (h *chanHandler) GetError() []*chan *agentinterfaces.ErrorResponse {
	return []*chan *agentinterfaces.ErrorResponse{&h.errorChan}
}

// GetUnhandled returns the unhandled event channels.
func (h *chanHandler) GetUnhandled() []*chan *[]byte {
	return []*chan *[]byte{&h.unhandledChan}
}

// Verify interface compliance at compile time.
var _ agentinterfaces.AgentMessageChan = (*chanHandler)(nil)
