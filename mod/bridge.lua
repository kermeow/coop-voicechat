gVoiceBridge = {}

local RECV_MOD_FS_NAME = "coop-voicechat-recv"

gVoiceBridge.active = false

gVoiceBridge.sendFS = mod_fs_get()
gVoiceBridge.recvFS = mod_fs_get(RECV_MOD_FS_NAME)

-- checks if new data is available
local function bridge_poll()
    gVoiceBridge.recvFS = mod_fs_reload(RECV_MOD_FS_NAME)
    if not gVoiceBridge.recvFS then
        return false
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
    end
    bridge_send()
end

hook_event(HOOK_UPDATE, bridge_update)
