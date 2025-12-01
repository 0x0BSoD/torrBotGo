package transmission

import (
	"github.com/0x0BSoD/transmission"
)

type Client struct {
	API     *transmission.Client
	Storage struct {
		Torrent    *transmission.Torrent
		tFile      []byte
		magentLink string
		messageID  int
	}
}
