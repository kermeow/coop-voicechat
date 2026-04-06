---@param modfs ModFs
---@param filepath string
---@param text boolean
---@return ModFsFile
function mod_fs_get_or_create_file(modfs, filepath, text)
    return modfs:get_file(filepath) or modfs:create_file(filepath, text)
end