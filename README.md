# livebuffer
A Go-based tool for buffering twitch livestreams and providing flexible access to stream clips via a REST API.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Setup

### Docker (recommended)
The easiest way to run the tool is using Docker. A pre-built image is available on the [GitHub Container Registry](https://github.com/matthiasharzer/livebuffer/pkgs/container/livebuffer).

#### Docker Compose
Create a `docker-compose.yml` file and start it with `docker compose up -d`. Make sure to adjust the command parameters as needed.

```yaml
services:
	livebuffer:
		image: ghcr.io/matthiasharzer/livebuffer:latest
		container_name: livebuffer
		restart: unless-stopped
		ports:
			- "4000:4000"
		environment:
			TWITCH_CLIENT_ID: <your_client_id>
			TWITCH_CLIENT_SECRET: <your_client_secret>

		command: twitch run --port 4000 --username <twitch_username> --public-url <https://yourdomain.com>
```
> [!Note]
> Make sure to replace `<your_client_id>`, `<your_client_secret>`, `<twitch_username>`, and `<https://yourdomain.com>` with your actual Twitch API credentials, the username of the livestream you want to buffer, and the public URL where the REST API will be accessible.

Quick reference for the command parameters:
- `--port`: The port on which the REST API will be available (default: 4000).
- `--username`: The Twitch username of the livestream to buffer.
- `--public-url`: The public URL where the REST API will be accessible (used for twitch webhooks).


#### Docker CLI
```bash
docker run -d \
	--name livebuffer \
	-p 4000:4000 \
	ghcr.io/matthiasharzer/livebuffer:latest \
	twitch run --port 4000 --username <twitch_username> --public-url <https://yourdomain.com>
```

### Binary
Download the [latest release](https://github.com/matthiasharzer/livebuffer/releases/latest) for your platform and run it with the appropriate command-line arguments.

## Usage

### `twitch run` Command
```bash
./livebuffer twitch run --port 4000 --public-url <https://yourdomain.com> --username <twitch_username>
```

#### Command-Line Flags

| Flag                | Required | Default               | Description                                                                                                                                                                                                               |
|---------------------|----------|-----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--username` / `-u` | ✅       |                       | The Twitch username for which to buffer the livestream.                                                                                                                                                                   |
| `--public-url`      | ✅       |                       | The public base URL for the REST API. Used for Twitch EventSub Webhook notifications, which require a public accesible URL.                                                                                               |
| `--port` / `-p`     | ❌       | 4000                  | The port on which the REST API will be available.                                                                                                                                                                         |
| `--host`            | ❌       | `""` (all interfaces) | The host/IP address on which the REST API will listen.                                                                                                                                                                    |
| `--buffer-dir`      | ❌       | _temporary directory_ | The directory where the livestream buffer will be stored. By default, a temporary directory will be created and used. If you want to persist the buffer across restarts, specify a directory here (e.g., `/data/buffer`). |
| `--max-streams`     | ❌       | 2                     | The maximum number of streams to keep on disk.                                                                                                                                                                            |

#### Environment Variables

The following environment variables *must* be set for Twitch API authentication:
- `TWITCH_CLIENT_ID`: Your Twitch API client ID.
- `TWITCH_CLIENT_SECRET`: Your Twitch API client secret.

> Twitch API credentials can be obtained by registering an application on the [Twitch Developer Console](https://dev.twitch.tv/console/apps).

#### API Endpoints
| Method | Endpoint                                                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|--------|------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| GET    | `/api/v1/list`                                                         | Lists all available livestreams (archived and live).                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| GET    | `/api/v1/download?stream_id=<stream_id>`                               | Downloads a specific livestream by its ID as a file.                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| GET    | `/api/v1/clip?stream_id=<stream_id>&start=<start_time>&end=<end_time>` | Creates a clip from a specific livestream by providing the start and end times. <br/>`start` and `end` must be in a format parsable by [Go's `time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) (e.g., `10m` for 10 minutes). Valid units are "ns", "us" (or "µs"), "ms", "s", "m", "h". <br/> If `end` is greater than the stream's duration, the clip will be truncated to the end of the stream. <br/>If `start` is greater than the stream's duration, and empty file will be returned. |


### `version` Command
Print the version of the tool:
```bash
./livebuffer version
```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
