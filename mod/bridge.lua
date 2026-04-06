local BRIDGE_VERSION = 1
local last_version_warning = 0

local RECV_MOD_FS_NAME = "coop-voicechat-recv"

do -- bridge_init
    mod_fs_hide_errors(true)

    gVoiceBridge = {}

    gVoiceBridge.connected = false

    gVoiceBridge.sendFS = mod_fs_get()
    gVoiceBridge.recvFS = mod_fs_get(RECV_MOD_FS_NAME)

    gVoiceBridge.sendFS:clear()

    gVoiceBridge.syncFile = gVoiceBridge.sendFS:create_file("sync", false)
    gVoiceBridge.syncLocalFrame = 1
    gVoiceBridge.syncRemoteFrame = 0
    gVoiceBridge.syncRemoteAckFrame = 0
    gVoiceBridge.syncLastRemoteFrame = 0
    gVoiceBridge.syncTimeoutCounter = 0
end

local function bridge_connect()
    gVoiceBridge.connected = true
    djui_popup_create("\\#60f060\\Voice Chat client connected!", 1)
end

local function bridge_disconnect()
    gVoiceBridge.connected = false
    djui_popup_create("\\#ffa060\\Voice Chat client disconnected!", 1)
end

-- checks if new data is available
local function bridge_poll()
    gVoiceBridge.recvFS = mod_fs_reload(RECV_MOD_FS_NAME)
    if not gVoiceBridge.recvFS then
        return false
    end

    local syncFile = gVoiceBridge.recvFS:get_file("sync")
    if not syncFile then
        return false
    end

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
    audio_send()
end

local function bridge_update()
    if bridge_poll() then
        bridge_recv()
        bridge_send()
    end

    gVoiceBridge.syncFile:rewind()
    gVoiceBridge.syncFile:write_integer(BRIDGE_VERSION, INT_TYPE_U16)
    gVoiceBridge.syncFile:write_integer(gVoiceBridge.syncLocalFrame, INT_TYPE_U32)
    gVoiceBridge.syncFile:write_integer(gVoiceBridge.syncRemoteFrame, INT_TYPE_U32)

    gVoiceBridge.sendFS:save()
end

hook_event(HOOK_UPDATE, bridge_update)
