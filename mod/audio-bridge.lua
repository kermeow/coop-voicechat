local FILE_HEADER = "smvc"

local PACK_PREFIX = "!1<"
local PACKET_MAX_SIZE = 1400 -- try to keep under mtu to avoid splitting, not the actual max (3000)

local MAX_AUDIO_FRAMES = 100

gVoiceStates = {}
for i = 0, MAX_PLAYERS - 1 do
    gVoiceStates[i] = {
        audioFile = "stream_" .. tostring(i),
        audioFrames = {},

        volume = 1,
        loudness = -1
    }
end

local function send_audio_packet(packet)
    local globalIndex = network_global_index_from_local(0)
    local prefix = string.pack(PACK_PREFIX .. "B", globalIndex)
    network_send_bytestring(false, prefix .. packet)
end

local function recv_audio_packet(packet)
    local globalIndex = string.unpack(PACK_PREFIX .. "B", packet)
    local localIndex = network_local_index_from_global(globalIndex)

    local lVoiceState = gVoiceStates[localIndex]

    packet = string.sub(packet, 2)

    while #packet > 0 do
        local timestamp, len = string.unpack(PACK_PREFIX .. "I4I4", packet)
        packet = string.sub(packet, 9)

        local data = string.sub(packet, 1, len)
        table.insert(lVoiceState.audioFrames, { syncFrame = 0, timestamp = timestamp, data = data })

        packet = string.sub(packet, len + 1)
    end
end

function audio_recv()
    -- get audio data from client
    local stream = gVoiceBridge.recvFs:get_file("stream")
    if not stream then
        return
    end

    stream:seek(4, FILE_SEEK_SET) -- skip file header

    local packet = ""
    ::read::
    while not stream:is_eof() do
        local syncFrame = stream:read_integer(INT_TYPE_U32)
        local timestamp = stream:read_integer(INT_TYPE_U32)
        local len = stream:read_integer(INT_TYPE_U32)
        local data = stream:read_bytes(len)
        if syncFrame <= gVoiceBridge.syncLastRemoteFrame then
            goto read
        end
        local segment = string.pack(PACK_PREFIX .. "I4I4", timestamp, len) .. data
        if #segment > PACKET_MAX_SIZE then
            -- frame is way too big, send separately
            send_audio_packet(segment)
            goto read
        end
        if #packet + #segment > PACKET_MAX_SIZE then
            send_audio_packet(packet)
            packet = segment
            goto read
        end
        packet = packet .. segment
    end

    if #packet > 0 then
        send_audio_packet(packet)
    end

    -- get loudness from client
    local loudness = gVoiceBridge.recvFs:get_file("loudness")
    if not loudness then
        return
    end

    loudness:seek(4, FILE_SEEK_SET)

    while not loudness:is_eof() do
        local i = loudness:read_integer(INT_TYPE_U8)
        local vol = loudness:read_number(FLOAT_TYPE_F64)
        gVoiceStates[i].loudness = math.max(0, amp2db(vol) + 100) / 100
    end
end

function audio_send()
    local i = 0

    -- send audio data to client
    ::write::
    while i < MAX_PLAYERS - 1 do
        i = i + 1

        local lVoiceState = gVoiceStates[i]
        local lNetworkPlayer = gNetworkPlayers[i]

        if not lNetworkPlayer.connected then
            if gVoiceBridge.sendFs:get_file(lVoiceState.audioFile) then
                gVoiceBridge.sendFs:delete_file(lVoiceState.audioFile)
            end
            goto write
        end

        local stream = mod_fs_get_or_create_file(gVoiceBridge.sendFs, lVoiceState.audioFile, false)
        mod_fs_file_clear(stream)
        stream:write_bytes(FILE_HEADER)

        table.sort(lVoiceState.audioFrames, function(a, b)
            return a.timestamp < b.timestamp
        end)

        local last = 0

        local deadFrames = math.max(0, #lVoiceState.audioFrames - MAX_AUDIO_FRAMES)

        for i, frame in pairs(lVoiceState.audioFrames) do
            if frame.syncFrame == 0 then
                frame.syncFrame = gVoiceBridge.syncLocalFrame
            end
            if frame.syncFrame < gVoiceBridge.syncRemoteAckFrame then
                deadFrames = math.max(deadFrames, i)
            else
                stream:write_integer(frame.syncFrame, INT_TYPE_U32)
                stream:write_integer(frame.timestamp, INT_TYPE_U32)
                stream:write_integer(#frame.data, INT_TYPE_U32)
                stream:write_bytes(frame.data)

                if last > 0 then
                    if last ~= frame.timestamp - 1 then
                        djui_chat_message_create(string.format("lost %d to %d", last, frame.timestamp))
                    end
                end
                last = frame.timestamp
            end
        end

        if deadFrames > 0 then
            for _ = 1, deadFrames do
                table.remove(lVoiceState.audioFrames, 1)
            end
        end
    end
end

-- why bytestring no network index? :(
local function on_bytestring_receive(raw)
    if not gVoiceBridge.connected then
        return
    end
    recv_audio_packet(raw)
end

hook_event(HOOK_ON_PACKET_BYTESTRING_RECEIVE, on_bytestring_receive)
