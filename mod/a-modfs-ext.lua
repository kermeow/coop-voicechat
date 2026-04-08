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

if not mod_fs_file_set_compression then
    -- pr not merged :(
    function mod_fs_file_set_compression(file, level)
        return true
    end
end
