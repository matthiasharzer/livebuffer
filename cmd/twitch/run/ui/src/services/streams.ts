interface StreamInfo {
	id: string;
	stream_state: 'live' | 'archived';
	size_bytes: number;
	duration_milliseconds: number;
}

interface StreamListResponse {
	streams: StreamInfo[];
}

const fetchStreamList = async (username: string): Promise<StreamInfo[]> => {
	if (!username) {
		return [];
	}
	const response = await fetch(`/api/v1/${username}/list`);
	if (!response.ok) {
		throw new Error(`Failed to fetch stream list: ${response.status} ${response.statusText}`);
	}
	const data: StreamListResponse = await response.json();
	return data.streams;
};

const fetchLiveStream = async (username: string): Promise<StreamInfo | null> => {
	const streams = await fetchStreamList(username);
	return streams.find(stream => stream.stream_state === 'live') || null;
};

const fetchStream = async (username: string, streamId: string): Promise<StreamInfo | null> => {
	if (!username || !streamId) {
		return null;
	}
	const streams = await fetchStreamList(username);
	return streams.find(stream => stream.id === streamId) || null;
};

export type { StreamInfo };
export { fetchLiveStream, fetchStream };
