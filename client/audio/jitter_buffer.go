package audio

const _MAX_TIMINGS = 40
const _MAX_TIMING_BUFFERS = 3
const _TOP_DELAY = 40

type _TimingBuffer struct {
	filled int
	count  int
	timing [_MAX_TIMINGS]int
	counts [_MAX_TIMINGS]int
}

func makeTimingBuffer() *_TimingBuffer {
	t := &_TimingBuffer{}
	t.filled = 0
	t.count = 0
	return t
}

func (t *_TimingBuffer) add(timing int) {
	var pos int
	if t.filled >= _MAX_TIMINGS && timing >= t.timing[t.filled-1] {
		t.count++
		return
	}

	for pos = 0; pos < t.filled && timing >= t.timing[pos]; pos++ {
	}

	if pos < t.filled {
		move := t.filled - pos
		if t.filled == _MAX_TIMINGS {
			move--
		}
		copy(t.timing[pos+1:], t.timing[pos:pos+move])
		copy(t.counts[pos+1:], t.counts[pos:pos+move])
	}

	t.timing[pos] = timing
	t.counts[pos] = t.count

	t.count++
	if t.filled < _MAX_TIMINGS {
		t.filled++
	}
}

const JITTER_MAX_BUFFER_SIZE = 200
const JITTER_MAX_LATE_RATE = 4
const JITTER_STEP_SIZE = 10

type JitterResult int

const (
	JitterOk JitterResult = iota
	JitterMissing
	JitterInsertion
)

type JitterPacket struct {
	Data      []byte
	Timestamp int
	Span      int
	Arrival   int
}

type JitterBuffer struct {
	packets      [JITTER_MAX_BUFFER_SIZE]*JitterPacket
	ptrTimestamp int
	nextStop     int
	lastReturned int
	buffered     int

	bufferMargin  int
	lostCount     int
	interpRequest int
	resetState    bool

	timings       [_MAX_TIMING_BUFFERS]*_TimingBuffer
	autoTradeoff  int
	windowSize    int
	subwindowSize int
}

func MakeJitterBuffer() *JitterBuffer {
	b := &JitterBuffer{
		ptrTimestamp: 0,
		nextStop:     0,
		lastReturned: 0,
		buffered:     0,

		bufferMargin:  0,
		lostCount:     0,
		interpRequest: 0,
		resetState:    true,
	}
	b.autoTradeoff = 32000
	b.windowSize = 100 * _TOP_DELAY / JITTER_MAX_LATE_RATE
	b.subwindowSize = b.windowSize / _MAX_TIMING_BUFFERS
	b.reset()
	return b
}

func (b *JitterBuffer) reset() {
	for i := 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
		p := b.packets[i]
		if p.Data != nil {
			p.Data = nil
		}
	}

	b.ptrTimestamp = 0
	b.nextStop = 0
	b.buffered = 0

	b.lostCount = 0
	b.interpRequest = 0
	b.resetState = true
	b.autoTradeoff = 32000
}

func (b *JitterBuffer) computeOptDelay() int {
	opt, late := 0, 0
	best, bestCost := 0, 0
	worst := 0
	penalty := false

	totalCount := 0
	for _, t := range b.timings {
		totalCount += t.count
	}
	if totalCount == 0 {
		return 0
	}

	// compute cost for one loss
	lateFactor := b.autoTradeoff * b.windowSize / totalCount

	pos := []int{0, 0, 0}

	// pick most late packets
	for i := 0; i < _TOP_DELAY; i++ {
		next, latest := -1, 32767
		for j := 0; j < _MAX_TIMING_BUFFERS; j++ {
			t := b.timings[j]
			if pos[j] < t.filled && t.timing[pos[j]] < latest {
				next = j
				latest = t.timing[pos[j]]
			}
		}
		if next != -1 {
			cost := 0
			if i == 0 {
				worst = latest
			}
			best = latest
			latest = min(latest, JITTER_STEP_SIZE)
			pos[next]++

			cost = -latest + lateFactor*late
			if cost < bestCost {
				bestCost = cost
				opt = latest
			}
		} else {
			break
		}

		// for next timing consider one more late
		late++
		// penalty for increasing number of late frames
		if latest >= 0 && !penalty {
			penalty = true
			late += 4
		}
	}

	deltaT := best - worst
	b.autoTradeoff = 1 + deltaT/_TOP_DELAY

	// dont reduce buffer size if we dont have much data
	if totalCount < _TOP_DELAY && opt > 0 {
		return 0
	}
	return opt
}

func (b *JitterBuffer) putTiming(timing int) {
	timing = max(timing, -32767)
	timing = min(timing, 32767)

	// rotate timing buffers if full
	if b.timings[0].count > b.subwindowSize {
		tmp := b.timings[_MAX_TIMING_BUFFERS-1]
		for i := _MAX_TIMING_BUFFERS - 1; i >= 1; i-- {
			b.timings[i] = b.timings[i-1]
		}
		b.timings[0] = tmp
		b.timings[0].filled = 0
		b.timings[0].count = 0
	}
	b.timings[0].add(timing)
}

func (b *JitterBuffer) shiftTimings(amount int) {
	for i := 0; i < _MAX_TIMING_BUFFERS; i++ {
		t := b.timings[i]
		for j := 0; j < t.filled; j++ {
			t.timing[j] += amount
		}
	}
}

func (b *JitterBuffer) Put(p *JitterPacket) {
	var i int
	late := false

	// remove old packets
	if b.resetState {
		for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
			pi := b.packets[i]
			if pi.Data != nil && p.Timestamp+p.Span <= b.ptrTimestamp {
				pi.Data = nil
			}
		}
	}

	if !b.resetState && p.Timestamp < b.nextStop {
		b.putTiming(p.Timestamp - b.nextStop - b.bufferMargin)
		late = true
	}

	if b.lostCount > 20 {
		// SON
		b.reset()
	}

	if b.resetState || p.Timestamp+p.Span >= b.ptrTimestamp {
		// find empty slot
		for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
			pi := b.packets[i]
			if pi.Data == nil {
				break
			}
		}

		// couldn't find empty, discard oldest packet
		if i == JITTER_MAX_BUFFER_SIZE {
			earliest := b.packets[0].Timestamp
			i = 0
			for j := 1; j < JITTER_MAX_BUFFER_SIZE; j++ {
				pj := b.packets[j]
				if pj.Data == nil || pj.Timestamp < earliest {
					earliest = pj.Timestamp
					i = j
				}
			}
			b.packets[i].Data = nil
		}

		in := b.packets[i]
		in.Data = make([]byte, len(p.Data))
		copy(in.Data, p.Data)
		in.Timestamp = p.Timestamp
		in.Span = p.Span
		if b.resetState || late {
			in.Arrival = 0
		} else {
			in.Arrival = b.nextStop
		}
	}
}

func (b *JitterBuffer) Get(out *JitterPacket, desiredSpan int, offset *int) JitterResult {
	var i int

	if b.resetState {
		found := false
		oldest := 0
		for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
			p := b.packets[i]
			if p.Data != nil && (!found || p.Timestamp < oldest) {
				oldest = p.Timestamp
				found = true
			}
		}
		if found {
			b.resetState = false
			b.ptrTimestamp = oldest
			b.nextStop = oldest
		} else {
			out.Timestamp = 0
			out.Span = b.interpRequest
			return JitterMissing
		}
	}

	b.lastReturned = b.ptrTimestamp

	if b.interpRequest != 0 {
		out.Timestamp = b.ptrTimestamp
		out.Span = b.interpRequest

		b.ptrTimestamp += b.interpRequest

		b.interpRequest = 0
		b.buffered = b.interpRequest - desiredSpan

		return JitterInsertion
	}

	// search for packet at the correct time with the correct length
	for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
		p := b.packets[i]
		if p.Data != nil && p.Timestamp == b.ptrTimestamp &&
			p.Timestamp+p.Span >= b.ptrTimestamp+desiredSpan {
			break
		}
	}

	// no match, search for packet at an older time with the needed length
	if i == JITTER_MAX_BUFFER_SIZE {
		for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
			p := b.packets[i]
			if p.Data != nil && p.Timestamp <= b.ptrTimestamp &&
				p.Timestamp+p.Span >= b.ptrTimestamp+desiredSpan {
				break
			}
		}
	}

	// still no match, search for packet at an older time with part of the needed length
	if i == JITTER_MAX_BUFFER_SIZE {
		for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
			p := b.packets[i]
			if p.Data != nil && p.Timestamp <= b.ptrTimestamp &&
				p.Timestamp+p.Span >= b.ptrTimestamp {
				break
			}
		}
	}

	// its over, get the earliest packet
	if i == JITTER_MAX_BUFFER_SIZE {
		found := false
		best, bestTime, bestSpan := 0, 0, 0
		for i = 0; i < JITTER_MAX_BUFFER_SIZE; i++ {
			// check if packet is within desired span
			p := b.packets[i]
			if p.Data != nil && p.Timestamp >= b.ptrTimestamp &&
				p.Timestamp < b.ptrTimestamp+desiredSpan {
				if !found || p.Timestamp < bestTime || p.Span > bestSpan {
					found = true
					best, bestTime, bestSpan = i, p.Timestamp, p.Span
				}
			}
		}
		if found {
			i = best
		}
	}

	// we got something
	if i != JITTER_MAX_BUFFER_SIZE {
		p := b.packets[i]

		b.lostCount = 0

		if p.Arrival != 0 {
			b.putTiming(p.Timestamp - p.Arrival - b.bufferMargin)
		}

		*offset = p.Timestamp - b.ptrTimestamp
		b.buffered = p.Span - desiredSpan

		out.Data = make([]byte, len(p.Data))
		copy(out.Data, p.Data)
		out.Timestamp = p.Timestamp
		out.Span = p.Span

		b.lastReturned = p.Timestamp
		b.ptrTimestamp = p.Timestamp + p.Span

		return JitterOk
	}

	// we have no packet noo
	b.lostCount++

	opt := b.computeOptDelay()

	if opt < 0 {
		// need to increase buffering
		b.shiftTimings(-opt)
		out.Timestamp = b.ptrTimestamp
		out.Span = -opt
		b.buffered = out.Span - desiredSpan
		return JitterInsertion
	} else {
		// normal packet loss
		out.Timestamp = b.ptrTimestamp
		out.Span = min(desiredSpan, JITTER_STEP_SIZE)
		b.ptrTimestamp += out.Span
		b.buffered = 0
		return JitterMissing
	}
}
