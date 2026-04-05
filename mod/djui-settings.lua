-- Util render funcs --

---@param x integer
---@param y integer
---@param width integer
---@param height integer
local function djui_hud_render_djui(x, y, width, height)
    local sDjuiTheme = djui_menu_get_theme()
    local rectColor = sDjuiTheme.threePanels.rectColor
    local borderColor = sDjuiTheme.threePanels.borderColor
    djui_hud_set_color(borderColor.r, borderColor.g, borderColor.b, borderColor.a)
    djui_hud_render_rect(x, y, 8, height)
    djui_hud_render_rect(x + width - 8, y, 8, height)
    djui_hud_render_rect(x + 8, y, width - 16, 8)
    djui_hud_render_rect(x + 8, y + height - 8, width - 16, 8)
    djui_hud_set_color(rectColor.r, rectColor.g, rectColor.b, rectColor.a)
    djui_hud_render_rect(x + 8, y + 8, width - 16, height - 16)
end

-- Djui stuffs
local FONT_USER = djui_menu_get_font()

currHeldSlider = MAX_PLAYERS
local function render_user_volume_slider(index, x, y, width, height)
    local v = gVoiceStates[index]

    -- WIP var for user speaking volume being displayed
    v.speakVol = (math.sin(get_global_timer()*0.1 + index*0.3)/math.pi) + 0.5

    local volumeScale = width*v.volume*0.5

    -- Slider Input
    if currHeldSlider == MAX_PLAYERS or currHeldSlider == index then
        local mX = djui_hud_get_mouse_x()
        local mY = djui_hud_get_mouse_y()
        if currHeldSlider ~= index then
            -- Check if mouse is over slider and held
            if x <= mX and mX <= x + width and y <= mY and mY <= y + height and djui_hud_get_mouse_buttons_down() & (MOUSE_BUTTON_1 | MOUSE_BUTTON_3) ~= 0 then
                currHeldSlider = index
                if djui_hud_get_mouse_buttons_down() & MOUSE_BUTTON_3 ~= 0 then
                    v.volume = v.volume == 0 and 1 or 0
                end
            end
        else
            if djui_hud_get_mouse_buttons_down() & MOUSE_BUTTON_1 ~= 0 then
                v.volume = math.clamp(mX - x, 0, width)/width*2
            end
        end
    end

    -- Render User Info
    djui_hud_set_font(FONT_USER)
    local textName = gNetworkPlayers[index].name .. (index == 0 and " (Input)" or "")
    djui_hud_print_text(textName, x, y + height, 1)
    local textVol = tostring(math.round(v.volume*100)).."%"
    djui_hud_print_text(textVol, x + width - djui_hud_measure_text(textVol), y + height, 1)

    -- Volume Bars
    djui_hud_set_color(255, 255, 255, 100)
    djui_hud_render_rect(x, y, width, height)
    djui_hud_render_rect(x, y, volumeScale*v.speakVol, height)

    -- Volume Slider
    djui_hud_set_color(255, 255, 255, 255)
    djui_hud_render_rect(x + volumeScale - 5, y - 2, 10, height + 4)

    v.volume = math.clamp(v.volume, 0, 2)
end

local function render_voice_settings()
    if not djui_hud_is_pause_menu_created() then return end

    djui_hud_set_resolution(RESOLUTION_DJUI)
    local screenWidth = djui_hud_get_screen_width()
    local screenHeight = djui_hud_get_screen_height()
    djui_hud_render_djui(screenWidth - 500, 0, 500, screenHeight)

    djui_hud_set_color(255, 255, 255, 255)
    local userHeight = (screenHeight - 110)/(MAX_PLAYERS)
    render_user_volume_slider(0, screenWidth - 450, screenHeight - 20 - (userHeight*1.5), 400, math.max(userHeight*1.5 - 35, 5))
    local inactive = 0
    for i = 1, MAX_PLAYERS - 1 do
        if true then--gNetworkPlayers[i].connected then
            render_user_volume_slider(i, screenWidth - 400, screenHeight - 20 - (userHeight*1.5) - ((userHeight + 2)*(i - inactive)), 350, math.max(userHeight - 35, 5))
        else
            inactive = inactive + 1
        end
    end

    -- Reset held slider if not held
    if djui_hud_get_mouse_buttons_down() & (MOUSE_BUTTON_1 | MOUSE_BUTTON_3) == 0 then
        currHeldSlider = MAX_PLAYERS
    end
end

local function hud_render()
    FONT_USER = djui_menu_get_font()
    render_voice_settings()
end

hook_event(HOOK_ON_HUD_RENDER, hud_render)