# Transmission telegram bot

![Docker Automated build](https://img.shields.io/docker/automated/zlodey23/transmission-bot)
![Build](https://github.com/0x0BSoD/torrBotGo/workflows/Build/badge.svg)
![Release](https://github.com/0x0BSoD/torrBotGo/workflows/Release/badge.svg)

Telegram bot as a transmission interface.

#### Features
 * Add torrent by file / magnet link
 * Get short info about status
 * Start / Stop torrents
 * Change torrent priority
 * Remove torrent
 * Message when torrent done
 * Can get image of the torrent from rutracker if used torrent file
 * Search torrent by name with prefix *t:* i.e *t:Ubuntu* 
 
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

## Docker

ImgDir can't be `/tmp`. 

Create the config.json described above and run:

```
docker run -v "$PWD/path/to/config.json":/config.json zlodey23/transmission-bot  
```