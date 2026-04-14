---@param modfs ModFs
---@param filepath string
---@param text boolean
---@return ModFsFile
function mod_fs_get_or_create_file(modfs, filepath, text)
    return modfs:get_file(filepath) or modfs:create_file(filepath, text)
end

---@param file ModFsFile
---@return boolean
function mod_fs_file_clear(file)
    return file:rewind() and file:erase(file.size)
end

---@param file ModFsFile
---@param localIndex integer
---@return boolean
function mod_fs_file_write_player(file, localIndex)
    local lVoiceState = gVoiceStates[localIndex]
    local lMarioState = gMarioStates[localIndex]
    local lNetworkPlayer = gNetworkPlayers[localIndex]

    local pos = lMarioState.pos
    file:write_number(pos.x, FLOAT_TYPE_F64)
    file:write_number(pos.y + lMarioState.marioObj.hitboxHeight - 60, FLOAT_TYPE_F64)
    file:write_number(pos.z, FLOAT_TYPE_F64)

    file:write_integer(lNetworkPlayer.currLevelNum, INT_TYPE_U16)
    file:write_integer(lNetworkPlayer.currAreaIndex, INT_TYPE_U16)
    file:write_integer(lMarioState.currentRoom, INT_TYPE_U16)

    file:write_integer(lMarioState.cap, INT_TYPE_U8)
    file:write_integer(lMarioState.waterLevel, INT_TYPE_U16)

    file:write_number(lVoiceState.volume, FLOAT_TYPE_F64)

    return true
end
