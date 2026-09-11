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
- `--public-url`: The public URL where the REST API will be accessible (used for twitch webhooks). This should be the domain + protocol, not the full path to the API endpoint (e.g., `https://yourdomain.com`).


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

| Flag                | Required | Default               | Description                                                                                                                                                                                                                                              |
|---------------------|----------|-----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--username` / `-u` | ✅       |                       | The Twitch username for which to buffer the livestream. This argument can be used multiple times to monitor multiple users.                                                                                                                              |
| `--public-url`      | ✅       |                       | The public base URL for the REST API. Used for Twitch EventSub Webhook notifications, which require a public accessible URL. This is usually the endpoint for your reverse proxy which forwards requests to the REST API.                                |
| `--port` / `-p`     | ❌       | 4000                  | The port on which the REST API will be available.                                                                                                                                                                                                        |
| `--host`            | ❌       | `""` (all interfaces) | The host/IP address on which the REST API will listen.                                                                                                                                                                                                   |
| `--buffer-dir`      | ❌       | _temporary directory_ | The directory where the livestream buffer will be stored. By default, a temporary directory will be created and used. If you want to persist the buffer across restarts, specify a directory here (e.g., `/data/buffer`).                                |
| `--max-streams`     | ❌       | 2                     | The maximum number of streams to keep on disk.                                                                                                                                                                                                           |
| `--eventsub-secret` | ❌       | _random string_       | The secret used for Twitch EventSub Webhook notifications. Twitch uses this secret string to authenticate webhook notification against the REST APIs eventsub endpoint. Must be between 10 and 100 characters long and may only contain ASCII characters |

#### Environment Variables

The following environment variables *must* be set for Twitch API authentication:
- `TWITCH_CLIENT_ID`: Your Twitch API client ID.
- `TWITCH_CLIENT_SECRET`: Your Twitch API client secret.

> Twitch API credentials can be obtained by registering an application on the [Twitch Developer Console](https://dev.twitch.tv/console/apps).

#### API Endpoints
| Method | Endpoint                                                                          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
|--------|-----------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| POST   | `/api/v1/twitch-event-sub`                                                        | The enpoint for twitch webhook notifications.                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| GET    | `/api/v1/{username}/list`                                                         | Lists all available livestreams of the user (archived and live).                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| GET    | `/api/v1/{username}/download?stream_id=<stream_id>`                               | Downloads a specific livestream of the user by its ID as a file (MPEG-TS).                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| GET    | `/api/v1/{username}/clip?stream_id=<stream_id>&start=<start_time>&end=<end_time>` | Creates a clip from a specific livestream of the user by providing the start and end times (MPEG-TS). <br/>`start` and `end` must be in a format parsable by [Go's `time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) (e.g., `10m` for 10 minutes). Valid units are "ns", "us" (or "µs"), "ms", "s", "m", "h". <br/> If `end` is greater than the stream's duration, the clip will be truncated to the end of the stream. <br/>If `start` is greater than the stream's duration, an empty file will be returned. |
| GET    | `/api/v1/{username}/live/*`                                                       | Provides the HLS manifest and chunks of the current stream of the user, if they are live. The manifest can be found under `index.m3u8`. The chunks use the path `/api/v1/{username}/live/` as a prefix (e.g. `/api/v1/{username}/live/chunk_00001.ts`).                                                                                                                                                                                                                                                                    |
| GET    | `/api/v1/{username}/video/{streamID}/*`                                           | Similar to `/api/v1/{username}/live/*`, but provides the HLS manifest and chunks for the specified `streamID`.                                                                                                                                                                                                                                                                                                                                                                                                             |

> Note: The `username` is the broadcaster's Twitch login name (all lowercase)

#### UI
livebuffer comes with a small UI to restream buffered livestreams or recordings. Currently, the UI only supports the following endpoints:
- `/{username}/live`: Provides a video player for the current livestream of the user, if they are live.
- `/{username}/video?stream_id={streamID}`: Provides a video player for the specified `streamID`.

At the moment, it is not possible to browse the list of available streams/users or show/create clips via the UI. This is planned for a future release.

### `version` Command
Print the version of the tool:
```bash
./livebuffer version
```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
