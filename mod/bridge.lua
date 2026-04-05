local BRIDGE_VERSION = 1 -- todo: add version checking on connect

local RECV_MOD_FS_NAME = "coop-voicechat-recv"

do -- bridge_init
    mod_fs_hide_errors(true)

    gVoiceBridge = {}

    gVoiceBridge.active = false

    gVoiceBridge.sendFS = mod_fs_get()
    gVoiceBridge.recvFS = mod_fs_get(RECV_MOD_FS_NAME)

    gVoiceBridge.sendFS:clear()

    gVoiceBridge.syncFile = gVoiceBridge.sendFS:create_file("sync", false)
    gVoiceBridge.syncLocalFrame = 0
    gVoiceBridge.syncRemoteFrame = 0
    gVoiceBridge.syncRemoteAckFrame = 0
end

local function bridge_activate()
    gVoiceBridge.active = true
    djui_chat_message_create("Voice chat connected")
end

local function bridge_deactivate()
    gVoiceBridge.active = false
    djui_chat_message_create("Voice chat disconnected")
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

    local lastActive = gVoiceBridge.active
    local lastRemoteFrame = gVoiceBridge.syncRemoteFrame

    gVoiceBridge.syncRemoteFrame = syncFile:read_integer(INT_TYPE_U32)
    gVoiceBridge.syncRemoteAckFrame = syncFile:read_integer(INT_TYPE_U32)

    local ackFrameValid = gVoiceBridge.syncRemoteAckFrame <= gVoiceBridge.syncLocalFrame
    local ackFrameThreshold = gVoiceBridge.syncLocalFrame - gVoiceBridge.syncRemoteAckFrame < 3

    if gVoiceBridge.syncRemoteFrame > lastRemoteFrame then
        -- active means the client is running and acknowledging us
        local shouldActivate = ackFrameValid and ackFrameThreshold
        if shouldActivate and not lastActive then
            bridge_activate()
        end
        return gVoiceBridge.active
    end

    if lastActive and not (ackFrameValid and ackFrameThreshold) then
        bridge_deactivate()
    end

    return false
end

-- handles new data
local function bridge_recv()
end

-- sends new data
local function bridge_send()
end

local function bridge_update()
    if bridge_poll() then
        bridge_recv()
        bridge_send()
    end

    gVoiceBridge.syncLocalFrame = get_global_timer()

    gVoiceBridge.syncFile:rewind()
    gVoiceBridge.syncFile:write_integer(gVoiceBridge.syncLocalFrame, INT_TYPE_U32)
    gVoiceBridge.syncFile:write_integer(gVoiceBridge.syncRemoteFrame, INT_TYPE_U32)

    gVoiceBridge.sendFS:save()
end

hook_event(HOOK_UPDATE, bridge_update)
