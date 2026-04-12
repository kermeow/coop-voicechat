local BRIDGE_VERSION = 2
local last_version_warning = 0

local RECV_MOD_FS_NAME = "coop-voicechat-recv"
local FILE_HEADER = "smvc"

-- mod_fs_hide_errors(true)

gVoiceBridge = {}

gVoiceBridge.connected = false

gVoiceBridge.sendFs = mod_fs_get()
gVoiceBridge.recvFs = mod_fs_get(RECV_MOD_FS_NAME)

gVoiceBridge.sendFs:clear()

gVoiceBridge.syncFile = gVoiceBridge.sendFs:create_file("sync", false)
mod_fs_file_set_compression(gVoiceBridge.syncFile, 0)

gVoiceBridge.syncLocalFrame = 1
gVoiceBridge.syncRemoteFrame = 0
gVoiceBridge.syncRemoteAckFrame = 0
gVoiceBridge.syncLastRemoteFrame = 0
gVoiceBridge.syncTimeoutCounter = 0

local function bridge_connect()
    gVoiceBridge.connected = true
    play_sound(SOUND_GENERAL_COIN, gGlobalSoundSource)
    djui_popup_create("Voice Chat:\n\\#60f060\\Client Connected!", 2)
end

local function bridge_disconnect()
    gVoiceBridge.connected = false
    play_sound(SOUND_MENU_PAUSE_HIGHPRIO, gGlobalSoundSource)
    djui_popup_create("Voice Chat:\n\\#ffa060\\Client Disconnected!", 2)
end

-- checks if new data is available
local function bridge_poll()
    mod_fs_hide_errors(true)
    -- this is the most likely operation to fail
    gVoiceBridge.recvFs = mod_fs_reload(RECV_MOD_FS_NAME)
    mod_fs_hide_errors(false)
    if not (gVoiceBridge.sendFs and gVoiceBridge.recvFs) then
        return false
    end

    local syncFile = gVoiceBridge.recvFs:get_file("sync")
    if not syncFile then
        return false
    end

    syncFile:seek(4, FILE_SEEK_SET)

    local remoteVersion = syncFile:read_integer(INT_TYPE_U16)
    if remoteVersion ~= BRIDGE_VERSION then
        local next_version_warning = last_version_warning + 900
        if last_version_warning == 0 or get_global_timer() >= next_version_warning then
            last_version_warning = get_global_timer()
            djui_popup_create(
                string.format("\\#f06060\\Voice Chat version mismatch!\n\\#dcdcdc\\Mod version: %d\nClient version: %d",
                    BRIDGE_VERSION, remoteVersion), 3)
        end
        return false
    end

    local lastActive = gVoiceBridge.connected
    local lastRemoteFrame = gVoiceBridge.syncRemoteFrame

    gVoiceBridge.syncRemoteFrame = syncFile:read_integer(INT_TYPE_U32)
    gVoiceBridge.syncRemoteAckFrame = syncFile:read_integer(INT_TYPE_U32)

    local ackFrameValid = gVoiceBridge.syncRemoteAckFrame > 0 and
        gVoiceBridge.syncRemoteAckFrame <= gVoiceBridge.syncLocalFrame
    local ackFrameThreshold = gVoiceBridge.syncLocalFrame - gVoiceBridge.syncRemoteAckFrame < 6

    gVoiceBridge.syncLocalFrame = get_global_timer()
    gVoiceBridge.syncLastRemoteFrame = lastRemoteFrame

    if gVoiceBridge.syncRemoteFrame > lastRemoteFrame then
        -- active means the client is running and acknowledging us
        local shouldActivate = ackFrameValid and ackFrameThreshold
        if shouldActivate then
            gVoiceBridge.syncTimeoutCounter = 0
            if not lastActive then bridge_connect() end
        end
        return gVoiceBridge.connected
    end

    if lastActive and not (ackFrameValid and ackFrameThreshold) then
        gVoiceBridge.syncTimeoutCounter = gVoiceBridge.syncTimeoutCounter + 1
        if gVoiceBridge.syncTimeoutCounter > 6 then
            log_to_console(
                string.format("Bridge disconnected - av:%s aft:%s slf:%d srf:%d sraf:%d stc:%d", ackFrameValid,
                    ackFrameThreshold,
                    gVoiceBridge.syncLocalFrame, gVoiceBridge.syncRemoteFrame, gVoiceBridge.syncRemoteAckFrame,
                    gVoiceBridge.syncTimeoutCounter),
                CONSOLE_MESSAGE_WARNING)
            bridge_disconnect()
        end
    end

    return false
end

-- handles new data
local function bridge_recv()
    audio_recv()
end

-- sends new data
local function bridge_send()
    local f = mod_fs_get_or_create_file(gVoiceBridge.sendFs, "local_player", false)
    mod_fs_file_clear(f)
    f:write_bytes(FILE_HEADER)

    mod_fs_file_write_player(f, 0)

    f = mod_fs_get_or_create_file(gVoiceBridge.sendFs, "players", false)
    mod_fs_file_clear(f)
    f:write_bytes(FILE_HEADER)

    for i = 1, MAX_PLAYERS - 1 do
        local lVoiceState = gVoiceStates[i]
        local lNetworkPlayer = gNetworkPlayers[i]
        if lNetworkPlayer.connected then
            f:write_integer(i | 0x80, INT_TYPE_U8)
            f:write_integer(#lVoiceState.audioFile, INT_TYPE_U8)
            f:write_bytes(lVoiceState.audioFile)
            mod_fs_file_write_player(f, i)
        else
            f:write_integer(i, INT_TYPE_U8)
        end
    end

    audio_send()
end

local function bridge_update()
    if bridge_poll() then
        bridge_recv()
        bridge_send()
    end

    gVoiceBridge.syncFile:rewind()
    gVoiceBridge.syncFile:write_bytes(FILE_HEADER)
    gVoiceBridge.syncFile:write_integer(BRIDGE_VERSION, INT_TYPE_U16)
    gVoiceBridge.syncFile:write_integer(gVoiceBridge.syncLocalFrame, INT_TYPE_U32)
    gVoiceBridge.syncFile:write_integer(gVoiceBridge.syncRemoteFrame, INT_TYPE_U32)

    gVoiceBridge.sendFs:save()
end

hook_event(HOOK_UPDATE, bridge_update)
