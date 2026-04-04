mod_fs_hide_errors(true)

local send_modfs = mod_fs_get()
local recv_modfs = mod_fs_get("coop-voicechat-recv")

local syncId = 0
local serverFlushed = false

local function recvSync()
    local syncFile = recv_modfs:get_file("sync")
    if not syncFile then
        return
    end
    sync = syncFile:read_integer(INT_TYPE_U32)
    serverFlushed = sync > syncId
    syncId = sync 
end

local function sendSync()
    local syncFile = send_modfs:get_file("sync") or send_modfs:create_file("sync", false)
    syncFile:rewind()
    syncFile:write_integer(syncId, INT_TYPE_U32)
end

local function poll()
    recv_modfs = mod_fs_reload("coop-voicechat-recv")
    if recv_modfs == nil then return end
    recvSync()

    if serverFlushed then
        -- todo: send data
    end

    sendSync()
    send_modfs:save()
end

hook_event(HOOK_UPDATE, poll)