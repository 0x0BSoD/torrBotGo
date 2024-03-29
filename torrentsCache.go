package main

import (
	"github.com/0x0BSoD/transmission"
)

type torrents struct {
	Items     transmission.TorrentMap
	hashIDMap map[int]string
}

func initCache(tMap transmission.TorrentMap) torrents {
	ctx.Mutex.Lock()
	var result torrents
	result.hashIDMap = make(map[int]string)
	for hash, i := range tMap {
		result.hashIDMap[i.ID] = hash
	}

	result.Items = tMap
	ctx.Mutex.Unlock()

	return result
}

func (t *torrents) getHash(id int) string {
	ctx.Mutex.Lock()
	r := t.hashIDMap[id]
	ctx.Mutex.Unlock()

	return r
}

func (t *torrents) getByID(id int) *transmission.Torrent {
	ctx.Mutex.Lock()
	if hash, ok := t.hashIDMap[id]; ok {
		return t.Items[hash]
	}
	ctx.Mutex.Unlock()

	return nil
}

// Update torrents cache and if some torrents done return array otherwise return nil
func (t *torrents) update(tMap transmission.TorrentMap) []*transmission.Torrent {
	ctx.Mutex.Lock()
	var changed []*transmission.Torrent
	for hash, oldI := range t.Items {
		if i, ok := tMap[hash]; ok {
			if oldI.Status != i.Status {
				changed = append(changed, i)
			}
		}
	}

	t.hashIDMap = make(map[int]string)
	for hash, i := range tMap {
		t.hashIDMap[i.ID] = hash
	}

	t.Items = tMap
	ctx.Mutex.Unlock()

	return changed
}
