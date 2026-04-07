local MAGIC_NUMBER = "smvc"
local PACKET_HEADER_FMT = "!1<BI4B"

do
    gVoiceStates = {}
    for i = 0, MAX_PLAYERS - 1 do
        gVoiceStates[i] = {
            volume = 1,
            speakVol = -1,
            frames = {}
        }
    end
end

---@param i number
---@param file ModFsFile
local function write_state_to_file(i, file)
    local voiceState = gVoiceStates[i]
    local marioState = gMarioStates[i]
    local networkPlayer = gNetworkPlayers[i]

    file:write_number(voiceState.volume, FLOAT_TYPE_F32)

    local headHeight = marioState.marioObj.hitboxHeight - 60

    file:write_number(marioState.pos.x, FLOAT_TYPE_F64)
    file:write_number(marioState.pos.y + headHeight, FLOAT_TYPE_F64)
    file:write_number(marioState.pos.z, FLOAT_TYPE_F64)

    file:write_integer(networkPlayer.currLevelNum, INT_TYPE_U8)
    file:write_integer(networkPlayer.currAreaIndex, INT_TYPE_U8)

    if i == 0 then
        file:write_number(gLakituState.curPos.x, FLOAT_TYPE_F64)
        file:write_number(gLakituState.curPos.y, FLOAT_TYPE_F64)
        file:write_number(gLakituState.curPos.z, FLOAT_TYPE_F64)

        file:write_number(gLakituState.curFocus.x, FLOAT_TYPE_F64)
        file:write_number(gLakituState.curFocus.y, FLOAT_TYPE_F64)
        file:write_number(gLakituState.curFocus.z, FLOAT_TYPE_F64)
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
            -- on_bytestring_receive(raw)
            frames = frames + 1
            order = order + 1
        end
    end

    local volumes = gVoiceBridge.recvFS:get_file("volumes")
    if not volumes then
        return
    end
    volumes:rewind()

    local newVolumes = {}
    while not volumes:is_eof() do
        local i = volumes:read_integer(INT_TYPE_U8)
        local rms = volumes:read_number(FLOAT_TYPE_F32)
        newVolumes[i] = rms
    end
    for i, voiceState in pairs(gVoiceStates) do
        local volume = newVolumes[i]
        if not volume then
            volume = -1
        else
            volume = math.min(1, math.sqrt(volume) * 2)
        end
        voiceState.speakVol = volume
    end
end

function audio_send()
    gVoiceBridge.sendFS:delete_file("states")
    local states = gVoiceBridge.sendFS:create_file("states", false)

    gVoiceBridge.sendFS:delete_file("local")
    local localFile = gVoiceBridge.sendFS:create_file("local", false)
    write_state_to_file(0, localFile)

    for i = 1, MAX_PLAYERS - 1 do
        local voiceState = gVoiceStates[i]
        local fileName = string.format("voice-%d", i)
        if gNetworkPlayers[i].connected then
            states:write_integer(i, INT_TYPE_U8)
            states:write_integer(#fileName, INT_TYPE_U8)
            states:write_bytes(fileName)
            write_state_to_file(i, states)

            gVoiceBridge.sendFS:delete_file(fileName)
            local sendFile = gVoiceBridge.sendFS:create_file(fileName, false)

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
        else
            states:write_integer(i | 0x80, INT_TYPE_U8)
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
        local data = string.sub(packet, 7)
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
