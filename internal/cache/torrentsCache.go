package cache

import (
	"maps"
	"sync"

	"github.com/0x0BSoD/transmission"
)

// Torrents - List of torrents that sved for it's state watching
type Torrents struct {
	items    transmission.TorrentMap
	idToHash map[int]string
	mu       sync.RWMutex
}

// New — init cache and build index
func New(tMap transmission.TorrentMap) *Torrents {
	c := &Torrents{
		items:    make(transmission.TorrentMap, len(tMap)),
		idToHash: make(map[int]string, len(tMap)),
	}
	for hash, t := range tMap {
		c.items[hash] = t
		c.idToHash[t.ID] = hash
	}
	return c
}

// GetHash — get hash by numeric ID.
func (c *Torrents) GetHash(id int) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	h, ok := c.idToHash[id]
	return h, ok
}

// GetByID — get torrent by numeric ID.
func (c *Torrents) GetByID(id int) (*transmission.Torrent, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	hash, ok := c.idToHash[id]
	if !ok {
		return nil, false
	}
	t, ok := c.items[hash]
	return t, ok
}

// GetByHash - get torrent by it hash
func (c *Torrents) GetByHash(hash string) (*transmission.Torrent, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.items[hash]
	if !ok {
		return nil, false
	}
	return t, true
}

// Update — update cache and return torrent with changed Status.
func (c *Torrents) Update(next transmission.TorrentMap) []*transmission.Torrent {
	c.mu.Lock()
	defer c.mu.Unlock()

	var changed []*transmission.Torrent
	for hash, oldT := range c.items {
		if nt, ok := next[hash]; ok && nt.Status != oldT.Status {
			changed = append(changed, nt)
		}
	}

	c.items = make(transmission.TorrentMap, len(next))
	c.idToHash = make(map[int]string, len(next))
	for hash, t := range next {
		c.items[hash] = t
		c.idToHash[t.ID] = hash
	}

	return changed
}

// Snapshot — safe copy state
func (c *Torrents) Snapshot() (transmission.TorrentMap, map[int]string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	itemsCopy := make(transmission.TorrentMap, len(c.items))
	maps.Copy(itemsCopy, c.items)

	indexCopy := make(map[int]string, len(c.idToHash))
	maps.Copy(indexCopy, c.idToHash)

	return itemsCopy, indexCopy
}
