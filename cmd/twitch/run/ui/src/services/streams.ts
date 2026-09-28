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
	const response = await fetch(`/api/v1/list?username=${encodeURIComponent(username)}`);
	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`Failed to fetch stream list: ${errorText}`);
	}
	const data: StreamListResponse = await response.json();
	return data.streams;
};

const fetchLiveStream = async (username: string): Promise<StreamInfo | null> => {
	const streams = await fetchStreamList(username);
	return streams.find(stream => stream.stream_state === 'live') || null;
};

const fetchStream = async (streamId: string): Promise<StreamInfo | null> => {
	if (!streamId) {
		return null;
	}

	const response = await fetch(`/api/v1/details/${encodeURIComponent(streamId)}`);
	if (response.status === 404) {
		return null;
	}
	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`Failed to fetch stream details: ${errorText}`);
	}

	return await response.json();
};

export type { StreamInfo };
export { fetchLiveStream, fetchStream };
