const formatDurationParts = (milliseconds: number): string[] => {
	const totalSeconds = Math.floor(milliseconds / 1000);
	const hours = Math.floor(totalSeconds / 3600);
	const minutes = Math.floor((totalSeconds % 3600) / 60);
	const seconds = totalSeconds % 60;

	const parts: string[] = [];
	if (hours > 0) {
		parts.push(`${hours}h`);
	}
	if (minutes > 0 || hours > 0) {
		parts.push(`${minutes}m`);
	}
	parts.push(`${seconds}s`);

	return parts;
};

const formatDuration = (milliseconds: number): string => {
	const parts = formatDurationParts(milliseconds);
	return parts.join(' ');
};

export { formatDuration, formatDurationParts };
