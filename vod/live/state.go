package live

import (
	"sync"

	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/matthiasharzer/livebuffer/twitch"
)

type broadcasterUserName = string
type streamID = string

type State struct {
	liveStreams map[broadcasterUserName]streamID

	unsubscribers []observer.UnsubscribeFunc
	mu            sync.RWMutex
}

func NewState() *State {
	return &State{
		liveStreams:   make(map[broadcasterUserName]streamID),
		unsubscribers: nil,
		mu:            sync.RWMutex{},
	}
}

func (s *State) GetLiveStreamID(username string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	liveStreamID, ok := s.liveStreams[username]
	if !ok {
		return ""
	}
	return liveStreamID
}

func (s *State) IsLive(username string) bool {
	return s.GetLiveStreamID(username) != ""
}

func (s *State) handleOnlineStateChange(state twitch.StreamOnlineState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state.IsOnline {
		s.liveStreams[state.BroadcasterUserName] = state.StreamID
	} else {
		delete(s.liveStreams, state.BroadcasterUserName)
	}
}

func (s *State) Add(channel observer.ReadonlyChannel[twitch.StreamOnlineState]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	unsubscribe := channel.Subscribe(s.handleOnlineStateChange)
	s.unsubscribers = append(s.unsubscribers, unsubscribe)
}

func (s *State) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, unsubscribe := range s.unsubscribers {
		unsubscribe()
	}
	s.liveStreams = make(map[broadcasterUserName]streamID)
	s.unsubscribers = nil
	return nil
}
