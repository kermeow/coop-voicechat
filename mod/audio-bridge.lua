local MAGIC_NUMBER = "smvc"

local PACKET_HEADER_FMT = "!1< B I4 B"

do
    gVoiceStates = {}
    for i = 0, MAX_PLAYERS - 1 do
        gVoiceStates[i] = {
            volume = 1,
            speakVol = 1,
            frames = {}
        }
    end
end

function audio_recv()
    local recording = gVoiceBridge.recvFS:get_file("recording")
    if not recording then
        return
    end
    recording:rewind()

    local frames = 0
    local order = 0
    while not recording:is_eof() do
        local syncFrame = recording:read_integer(INT_TYPE_U32)
        local len = recording:read_integer(INT_TYPE_U32)
        local data = recording:read_bytes(len)
        if syncFrame > gVoiceBridge.syncLastRemoteFrame then
            local raw = MAGIC_NUMBER ..
                string.pack(PACKET_HEADER_FMT, network_global_index_from_local(0), gVoiceBridge.syncLocalFrame, order) ..
                data
            network_send_bytestring(false, raw)
            frames = frames + 1
            order = order + 1
        end
    end
end

function audio_send()
    gVoiceBridge.sendFS:delete_file("states")
    local states = gVoiceBridge.sendFS:create_file("states", false)

    for i = 1, MAX_PLAYERS - 1 do
        local voiceState = gVoiceStates[i]
        if gNetworkPlayers[i].connected then
            states:write_integer(i, INT_TYPE_U8)
            local sendFile = mod_fs_get_or_create_file(gVoiceBridge.sendFS, tostring(i), false)

            -- todo: sort the frames just in case :p
            for i = #voiceState.frames, 1, -1 do
                local frame = voiceState.frames[i]
                if frame.syncFrame > 0 and frame.syncFrame <= gVoiceBridge.syncRemoteAckFrame then
                    table.remove(voiceState.frames, i)
                end
            end
            for _, frame in pairs(voiceState.frames) do
                if frame.syncFrame == 0 then
                    frame.syncFrame = gVoiceBridge.syncLocalFrame
                end
                sendFile:write_integer(frame.syncFrame, INT_TYPE_U32)
                sendFile:write_integer(string.len(frame.data), INT_TYPE_U32)
                sendFile:write_bytes(frame.data)
            end
        end
    end
end

-- why bytestring no network index? :(
local function on_bytestring_receive(raw)
    if not gVoiceBridge.connected then
        return
    end

    local packet = string.match(raw, "^" .. MAGIC_NUMBER .. "(.*)")
    if packet then
        local globalIndex, frame, order = string.unpack(PACKET_HEADER_FMT, packet)
        local data = string.sub(packet, 6)
        local localIndex = network_local_index_from_global(globalIndex)

        local voiceState = gVoiceStates[localIndex]
        table.insert(voiceState.frames, {
            syncFrame = 0,
            frame = frame,
            order = order,
            data = data
        })
    end
end

hook_event(HOOK_ON_PACKET_BYTESTRING_RECEIVE, on_bytestring_receive)
