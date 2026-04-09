package audio

const SAMPLE_RATE = 48_000 // 48khz

const OPUS_FRAME_SIZE = 0.02 // 20 ms
const OPUS_FRAME_SAMPLES = SAMPLE_RATE * OPUS_FRAME_SIZE
const OPUS_BITRATE = 64_000 // 64kbps
