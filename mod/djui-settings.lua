local TEX_MIC = get_texture_info("smvc_mic")
local TEX_SND = get_texture_info("smvc_snd")

--[[
    TODO:
        - Toggles
            - Stereo Panning
            - Mute and Deafen
            - Effects Slider
        - API
            - Effect Functions
            - Hear Local User Globally
            - Mute Local User
]]


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
    djui_hud_render_rect_interpolated(x, y, 8, height, x, y, 8, height)
    djui_hud_render_rect_interpolated(x + width - 8, y, 8, height, x + width - 8, y, 8, height)
    djui_hud_render_rect_interpolated(x + 8, y, width - 16, 8, x + 8, y, width - 16, 8)
    djui_hud_render_rect_interpolated(x + 8, y + height - 8, width - 16, 8, x + 8, y + height - 8, width - 16, 8)
    djui_hud_set_color(rectColor.r, rectColor.g, rectColor.b, rectColor.a)
    djui_hud_render_rect_interpolated(x + 8, y + 8, width - 16, height - 16, x + 8, y + 8, width - 16, height - 16)
end

local currHeldWidget = ""
local function djui_hud_render_slider(sliderID, currPos, currFill, min, max, x, y, w, h, funcHeld, funcRelease)
    currFill = currFill or 1
    local prevPos = currPos
    local djuiColor = djui_hud_get_color()
    local cR = djuiColor.r/255
    local cG = djuiColor.g/255
    local cB = djuiColor.b/255
    local cA = djuiColor.a/255

    -- Slider Input
    if currHeldWidget == "" or currHeldWidget == sliderID then
        local mX = djui_hud_get_mouse_x()
        local mY = djui_hud_get_mouse_y()
        if currHeldWidget ~= sliderID then
            -- Check if mouse is over slider and held
            if x <= mX and mX <= x + w and y <= mY and mY <= y + h and djui_hud_get_mouse_buttons_down() & (MOUSE_BUTTON_1) ~= 0 then
                currHeldWidget = sliderID
            end
        else
            if djui_hud_get_mouse_buttons_down() & MOUSE_BUTTON_1 ~= 0 then
                currPos = math.lerp(min, max, math.clamp(mX - (x + 2), 0, (w - 4)) / (w - 4))
                if funcHeld ~= nil then
                    funcHeld(currPos)
                end
            else
                if funcRelease ~= nil then
                    funcRelease(currPos)
                end
            end
        end
    end

    local text = tostring(math.round(currPos*100) .. "%")
    local textW, textH = djui_hud_measure_text(text)
    local textScale = h/textH
    local textX = x + (w - textW*textScale)*0.5
    local textY = y + (h - textH*textScale)*0.5
    local slideW = (w - 4)*((currPos - min)/max)
    djui_hud_set_color(220*cR, 220*cG, 220*cB, 255*cA)
    djui_hud_print_text(text, textX, textY, textScale)
    djui_hud_set_color(110*cR, 110*cG, 110*cB, 255*cA)
    djui_hud_render_rect(x + 2, y + 2, slideW, (h - 4))
    djui_hud_set_color(220*cR, 220*cG, 220*cB, 255*cA)
    djui_hud_render_rect(x + 2, y + 2, slideW*currFill, (h - 4))
    djui_hud_set_color(173*cR, 173*cG, 173*cB, 255*cA)
    djui_hud_render_rect(x, y, w, 2)
    djui_hud_render_rect(x, y + h - 2, w, 2)
    djui_hud_render_rect(x, y, 2, h)
    djui_hud_render_rect(x + w - 2, y, 2, h)

    djui_hud_set_scissor(0, 0, (x + 3 + slideW), djui_hud_get_screen_height())
    djui_hud_set_color(0, 0, 0, 255*cA)
    djui_hud_print_text(text, textX, textY, textScale)
    djui_hud_reset_scissor()

    djui_hud_set_color(djuiColor.r, djuiColor.g, djuiColor.b, djuiColor.a)
    return currPos
end

local function djui_hud_render_toggle(sliderID, currState, x, y, w, h, funcHeld, funcRelease)
    local djuiColor = djui_hud_get_color()
    local cR = math.min(djuiColor.r/220, 1)
    local cG = math.min(djuiColor.g/220, 1)
    local cB = math.min(djuiColor.b/220, 1)
    local cA = math.min(djuiColor.a/255, 1)

    -- Slider Input
    if currHeldWidget == "" or currHeldWidget == sliderID then
        local mX = djui_hud_get_mouse_x()
        local mY = djui_hud_get_mouse_y()
        if currHeldWidget ~= sliderID then
            -- Check if mouse is over slider and held
            if x <= mX and mX <= x + w and y <= mY and mY <= y + h and djui_hud_get_mouse_buttons_down() & (MOUSE_BUTTON_1) ~= 0 then
                currState = not currState
                currHeldWidget = sliderID
            end
        else
            if djui_hud_get_mouse_buttons_down() & MOUSE_BUTTON_1 ~= 0 then
                if funcHeld ~= nil then
                    funcHeld(currState)
                end
            else
                if funcRelease ~= nil then
                    funcRelease(currState)
                end
            end
        end
    end

    if currState then
        djui_hud_set_color(220*cR, 220*cG, 220*cB, 255*cA)
        djui_hud_render_rect(x + 8, y + 8, w - 16, h - 16)
    end
    djui_hud_set_color(173*cR, 173*cG, 173*cB, 255*cA)
    djui_hud_render_rect(x, y, w, 2)
    djui_hud_render_rect(x, y + h - 2, w, 2)
    djui_hud_render_rect(x, y, 2, h)
    djui_hud_render_rect(x + w - 2, y, 2, h)

    djui_hud_set_color(djuiColor.r, djuiColor.g, djuiColor.b, djuiColor.a)
    return currState
end

-- Djui stuffs
local FONT_USER = djui_menu_get_font()

local function render_voice_settings()
    if not djui_hud_is_pause_menu_created() then return end

    djui_hud_set_resolution(RESOLUTION_DJUI)
    local screenWidth = djui_hud_get_screen_width()
    local screenHeight = 1080
    local boxScaleX = 500
    local boxScaleY = screenHeight
    local boxX = screenWidth - boxScaleX
    local boxY = 0
    local heightScale = djui_hud_get_screen_height()/screenHeight
    djui_hud_render_djui(boxX, boxY, boxScaleX, boxScaleY*heightScale)

    -- Header
    djui_hud_set_font(FONT_USER)
    djui_hud_set_color(220, 220, 220, 255)
    djui_hud_print_text_interpolated("Voice Chat", boxX + 30, boxY + 20, 2, boxX + 30, boxY + 20, 2)

    djui_hud_set_font(FONT_USER)
    if gVoiceBridge.connected then
        -- Local Options
        local _, nameH = djui_hud_measure_text("Noise Suppression")
        djui_hud_print_text("Noise Suppression", boxX + 25, (150 - 32 + nameH)*heightScale, heightScale)
        gVoiceBridge.settings.suppression = djui_hud_render_toggle("settingSuppression", gVoiceBridge.settings.suppression, boxX + boxScaleX - 25 - 32*heightScale, 150*heightScale, 32*heightScale, 32*heightScale)

        local _, nameH = djui_hud_measure_text("Stereo Panning")
        djui_hud_print_text("Stereo Panning", boxX + 25, (200 - 32 + nameH)*heightScale, heightScale)
        gVoiceBridge.settings.stereoPan = djui_hud_render_toggle("settingStereoPan", gVoiceBridge.settings.stereoPan, boxX + boxScaleX - 25 - 32*heightScale, 200*heightScale, 32*heightScale, 32*heightScale)

        --local _, nameH = djui_hud_measure_text("Stereo Panning")
        --djui_hud_print_text("Stereo Panning", boxX + 25, (150 - 32 + nameH)*heightScale, heightScale)
        gVoiceBridge.settings.effectStrength = djui_hud_render_slider("effectStrength", gVoiceBridge.settings.effectStrength, nil, 0, 1, boxX + 25, 300*heightScale, 450, 32*heightScale)

        -- Remote User Options
        local inactive = 0
        for i = 1, MAX_PLAYERS - 1 do
            if true then --gNetworkPlayers[i].connected then
                local name = gNetworkPlayers[i].name..":"
                local _, nameH = djui_hud_measure_text(name)
                djui_hud_print_text(name, boxX+25, (screenHeight - 100 - 34*(i - inactive) - 16 + nameH*0.5)*heightScale, heightScale)
                gVoiceStates[i].clientMute = djui_hud_render_toggle("user"..tostring(i).."Mute", gVoiceStates[i].clientMute, boxX + 225 - 34*heightScale, (screenHeight - 100 - 34*(i - inactive))*heightScale, 32*heightScale, 32*heightScale)
                djui_hud_render_texture_tile(TEX_MIC, boxX + 225 - 34*heightScale, (screenHeight - 100 - 34*(i - inactive))*heightScale, 2*heightScale, 2*heightScale, gVoiceStates[i].clientMute and 16 or 0, 0, 16, 16)
                gVoiceStates[i].volume = djui_hud_render_slider("user"..tostring(i).."Volume", gVoiceStates[i].volume, gVoiceStates[i].loudness, 0, 2, boxX + 225, (screenHeight - 100 - 34*(i - inactive))*heightScale, 250, 32*heightScale)
            else
                inactive = inactive + 1
            end
        end

        -- Local User Options
        gVoiceStates[0].clientMute = djui_hud_render_toggle("user0Mute", gVoiceStates[0].clientMute, boxX + 25, (screenHeight - 57)*heightScale, 32*heightScale, 32*heightScale)
        gVoiceStates[0].deafen = djui_hud_render_toggle("user0Mute", gVoiceStates[0].deafen, boxX + 25 + 34*heightScale, (screenHeight - 57)*heightScale, 32*heightScale, 32*heightScale)
        if gVoiceStates[0].deafen then
            gVoiceStates[0].clientMute = true
        end
        gVoiceStates[0].mute = gVoiceStates[0].clientMute

        djui_hud_render_texture_tile(TEX_MIC, boxX + 25, (screenHeight - 57)*heightScale, 2*heightScale, 2*heightScale, gVoiceStates[0].clientMute and 16 or 0, 0, 16, 16)
        djui_hud_render_texture_tile(TEX_SND, boxX + 25 + 34*heightScale, (screenHeight - 57)*heightScale, 2*heightScale, 2*heightScale, gVoiceStates[0].deafen and 16 or 0, 0, 16, 16)

        if gVoiceStates[0].clientMute or gVoiceStates[0].deafen then
            djui_hud_set_color(255, 0, 0, 255)
        else
            djui_hud_set_color(255, 255, 255, 255)
        end
        gVoiceStates[0].volume = djui_hud_render_slider("user0Volume", gVoiceStates[0].volume, gVoiceStates[0].loudness, 0, 2, boxX + 25 + 68*heightScale, (screenHeight - 57)*heightScale, 450 - 68*heightScale, 32*heightScale)
    else
        local textScale = 1.5
        local textW, textH = djui_hud_measure_text("Client not Connected")
        local textW, textH = textW*textScale, textH*textScale
        djui_hud_print_text_interpolated("Client not Connected", boxX + boxScaleX*0.5 - textW*0.5, boxY + boxScaleY*0.5 - textH*0.5, textScale, boxX + boxScaleX*0.5 - textW*0.5, boxY + boxScaleY*0.5 - textH*0.5, textScale)
    end

    -- Reset held slider if not held
    if djui_hud_get_mouse_buttons_down() & (MOUSE_BUTTON_1) == 0 then
        currHeldWidget = ""
    end
end

local function hud_render()
    FONT_USER = djui_menu_get_font()

    render_voice_settings()
    
end

hook_event(HOOK_ON_HUD_RENDER, hud_render)