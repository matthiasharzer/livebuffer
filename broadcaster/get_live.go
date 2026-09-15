package broadcaster

func BuildGetLiveStreamsFunc(monitors []*Monitor) func() map[string]string {

	return func() map[string]string {
		livestreams := make(map[string]string)
		for _, monitor := range monitors {
			liveStreamID := monitor.GetLiveStreamID()
			if liveStreamID != "" {
				livestreams[monitor.broadcasterUserName] = liveStreamID
			}
		}
		return livestreams
	}
}
