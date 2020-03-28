# Transmission telegram bot

Telegram bot as a transmission interface.

#### Features
 * Add torrent by file / magnet link
 * Get short info about status
 * Start / Stop torrents
 * Change torrent priority
 * Remove torrent
 * Message when torrent done

Config:
```json
{
  "token":"TELEGRAM_TOKEN",
  "debug": false,
  "transmission": {
    "uri": "http://IP||HOSTNAME:PORT/transmission/rpc",
    "user": "USER",
    "password": "PASSWORD"
  },
  "img_dir": "/app/img",
  "default_download_dir": "/mnt/downloads",
  "categories": {
    "Films": "/Films",
    "TVSeries": "/TVSeries",
    "TVAnime": "/TVAnime",
    "Soft": "/Soft",
    "Education": "/Education"
  }
}
```