local hud_microphone_tex = get_texture_info("hud-microphone")

local function render_hud_microphone_status()
    if (hud_get_value(HUD_DISPLAY_FLAGS) & HUD_DISPLAY_FLAG_CAMERA) == 0 or (hud_get_value(HUD_DISPLAY_FLAGS) & HUD_DISPLAY_FLAG_CAMERA_AND_POWER) == 0 then return end

    local x = djui_hud_get_screen_width() - 70
    local y = 205

    local cameraHudStatus = hud_get_value(HUD_DISPLAY_CAMERA_STATUS)
    if cameraHudStatus == CAM_STATUS_NONE then
        x = x + 32
    end

    local tileX = 0
    if gVoiceBridge.connected then
        tileX = 16
    end
    djui_hud_render_texture_tile(hud_microphone_tex, x, y, 1, 1, tileX, 0, 16, 16)
end

local function on_hud_render_behind()
    djui_hud_set_resolution(RESOLUTION_N64)

    if gNetworkPlayers[0].currActNum == 99 or gMarioStates[0].action == ACT_INTRO_CUTSCENE or hud_is_hidden() then return end

    if not obj_get_first_with_behavior_id(id_bhvActSelector) then
        render_hud_microphone_status()
    end
end

hook_event(HOOK_ON_HUD_RENDER_BEHIND, on_hud_render_behind)
