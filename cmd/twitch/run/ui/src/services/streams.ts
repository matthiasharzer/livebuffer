interface BroadcasterInfo {
	username: string;
	streams: StreamInfo[];
}

interface StreamInfo {
	id: string;
	state: 'live' | 'archived';
	size_bytes: number;
	size: string; // human-readable size
	duration_milliseconds: number;
	started_at: string; // RFC 3339 timestamp
	title: string;
	username: string;
}

interface StreamListResponse {
	streams: StreamInfo[];
}

interface BroadcasterInfoResponse {
	broadcasters: BroadcasterInfo[];
}

interface StreamListOptions {
	state: 'live' | 'archived';
}

const fetchStreamList = async (
	username: string,
	options?: Partial<StreamListOptions>,
): Promise<StreamInfo[]> => {
	if (!username) {
		return [];
	}

	let url = `/api/v1/list?username=${encodeURIComponent(username)}`;
	if (options?.state) {
		url += `&state=${encodeURIComponent(options.state)}`;
	}

	const response = await fetch(url);
	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`Failed to fetch stream list: ${errorText}`);
	}
	const data: StreamListResponse = await response.json();
	return data.streams;
};

const fetchLiveStream = async (username: string): Promise<StreamInfo | null> => {
	const streams = await fetchStreamList(username, { state: 'live' });
	return streams.length > 0 ? streams[0] : null;
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

const fetchBroadcasters = async (): Promise<BroadcasterInfo[]> => {
	const response = await fetch('/api/v1/broadcasters');
	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`Failed to fetch broadcasters: ${errorText}`);
	}
	const data: BroadcasterInfoResponse = await response.json();
	const broadcasters = data.broadcasters.toSorted((a, b) => a.username.localeCompare(b.username));
	return broadcasters;
};

export type { BroadcasterInfo, StreamInfo };
export { fetchBroadcasters, fetchLiveStream, fetchStream };
