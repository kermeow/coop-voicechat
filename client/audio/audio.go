package audio

const SAMPLE_RATE = 24000 // opus is really annoying about this :(
const CHANNELS = 1        // mono

const OPUS_BITRATE = 32000      // 32 kbps
const OPUS_FRAME_SIZE_MS = 0.02 // 20 ms frames