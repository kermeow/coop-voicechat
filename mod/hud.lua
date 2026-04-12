local TEX_MIC = get_texture_info("hud-microphone")

local function render_player_microphone_status_interpolated(index, prevX, prevY, prevScale, x, y, scale)
    local v = gVoiceStates[index]
    local connected = v.loudness > 0

    local rotation, pivotX, pivotY = djui_hud_get_rotation()

    local tileX = 0
    if connected then
        tileX = 16
    end

    local speakScale = connected and (math.max(v.loudness, 0.15) - 0.15)*(1/0.85) or 0
    djui_hud_set_rotation(rotation + 0x1000*math.sin(get_global_timer())*speakScale, 0.5, 0.5)
    djui_hud_render_texture_tile_interpolated(TEX_MIC, prevX, prevY, prevScale, prevScale, x, y, scale, scale, tileX, 0, 16, 16)

    -- Cleanup
    djui_hud_set_rotation(rotation, pivotX, pivotY)
end

local function render_player_microphone_status(index, x, y, scale)
    render_player_microphone_status_interpolated(index, x, y, scale, x, y, scale)
end

local function render_hud_microphone_status()
    if (hud_get_value(HUD_DISPLAY_FLAGS) & HUD_DISPLAY_FLAG_CAMERA) == 0 or (hud_get_value(HUD_DISPLAY_FLAGS) & HUD_DISPLAY_FLAG_CAMERA_AND_POWER) == 0 then return end

    local x = djui_hud_get_screen_width() - 70
    local y = 205

    local cameraHudStatus = hud_get_value(HUD_DISPLAY_CAMERA_STATUS)
    if cameraHudStatus == CAM_STATUS_NONE then
        x = x + 32
    end

    render_player_microphone_status(0, x, y, 1)
end

sStateExtras = {}
for i = 0, MAX_PLAYERS do
    sStateExtras[i] = {
        prevPos = {x = 0, y = 0, z = 0},
        prevScale = 0,
        inited = false,
    }
end

local function render_nametag_microphone_status(index, pos)
    local np = gNetworkPlayers[index]
    local out = {x = 0, y = 0, z = 0}
    --djui_hud_world_pos_to_screen_pos(pos, out)
    if (not djui_hud_world_pos_to_screen_pos(pos, out)) then
        return;
    end

    local scale = -300 / out.z * djui_hud_get_fov_coeff();
    local measure = djui_hud_measure_text(np.name) * scale * 0.5;
    out.y = out.y - 16 * scale;

    alpha = (index == 0 and 255 or math.min(np.fadeOpacity << 3, 255)) * math.clamp(4 - scale, 0, 1);

    local e = sStateExtras[index];
    if (not e.inited) then
        vec3f_copy(e.prevPos, out);
        e.prevScale = scale;
        e.inited = true;
    end

    render_player_microphone_status_interpolated(index, e.prevPos.x + measure, e.prevPos.y, e.prevScale*2, out.x + measure, out.y, scale*2)

    vec3f_copy(e.prevPos, out);
    e.prevScale = scale;
end

local function on_hud_render_behind()
    djui_hud_set_resolution(RESOLUTION_N64)

    if gNetworkPlayers[0].currActNum == 99 or gMarioStates[0].action == ACT_INTRO_CUTSCENE or hud_is_hidden() then return end

    if not obj_get_first_with_behavior_id(id_bhvActSelector) then
        render_hud_microphone_status()
    end
end

hook_event(HOOK_ON_HUD_RENDER_BEHIND, on_hud_render_behind)
hook_event(HOOK_ON_NAMETAGS_RENDER, render_nametag_microphone_status)