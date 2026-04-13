package audio

import "github.com/gopxl/beep/v2"

const SAMPLE_RATE = 48_000 // 48khz
var SAMPLE_RATE_BEEP = beep.SampleRate(SAMPLE_RATE)

const OPUS_FRAME_SIZE = 0.02 // 20 ms
const OPUS_FRAME_SAMPLES = SAMPLE_RATE * OPUS_FRAME_SIZE
const OPUS_BITRATE = 64_000 // 64kbps
