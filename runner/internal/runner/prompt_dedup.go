package runner

type promptDedupRing struct {
	ids   []string
	index int
	count int
	cap   int
}

func newPromptDedupRing(capacity int) *promptDedupRing {
	if capacity <= 0 {
		capacity = 32
	}
	return &promptDedupRing{ids: make([]string, capacity), cap: capacity}
}

func (r *promptDedupRing) seen(id string) bool {
	for i := 0; i < r.count; i++ {
		if r.ids[i] == id {
			return true
		}
	}
	return false
}

func (r *promptDedupRing) add(id string) {
	if r.count < r.cap {
		r.ids[r.count] = id
		r.count++
		return
	}
	r.ids[r.index] = id
	r.index = (r.index + 1) % r.cap
}
