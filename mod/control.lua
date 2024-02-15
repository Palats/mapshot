local generated = require("generated")
local overrides = require("overrides")
local entities = require("entities")
local hash = require("hash")

local factorio_min_zoom = 0.031250

local all_surfaces = "_all_"

-- Read all settings and update the params var, incl. overrides.
function build_params(player)
  local params = {}
  -- settings.player[xxx] does contain the value at the beginning of the game,
  -- while get_player_settings contains the current value.
  local s = settings.get_player_settings(player)
  for k, v in pairs(s) do
    params[k] = v.value
  end

  if (params.surface == nil or params.surface == "") then
    params.surface = all_surfaces
  end

  for k,v in pairs(game.json_to_table(overrides)) do
    params[k] = v
  end

  if (string.sub(params.prefix, -1) ~= "/") then
    params.prefix = params.prefix .. "/"
  end

  params.tilemin = factorio_fit_zoom(params.resolution, params.tilemin, "tilemin")
  params.tilemax = factorio_fit_zoom(params.resolution, params.tilemax, "tilemax")

  return params
end

-- Calculate the zoom value for Factorio take_screenshot function
function factorio_zoom(render_size, tile_size)
  -- We want to have render_size pixels represent tile_size world unit.
  -- A zoom of 1.0 means that 32 pixels represent 1 world unit. A zoom of 2.0 means 64 pixels per world unit.
  return render_size / 32 / tile_size
end

function factorio_fit_zoom(render_size, tile_size, name)
  if (factorio_zoom(render_size, tile_size) < factorio_min_zoom) then
    local old = tile_size
    tile_size = render_size / 32 / factorio_min_zoom
    local msg = "Parameter " .. name .. " changed from " .. old .. " to " .. tile_size .. " to fit within Factorio minimal zoom of " .. factorio_min_zoom
    game.print(msg)
    log(msg)
  end
  return tile_size
end

-- Generate a full map screenshot.
function mapshot(params)
  log("mapshot params:\n" .. serpent.block(params))

  local unique_id = gen_unique_id()
  local map_id = gen_map_id()
  local savename = params.savename
  if (savename == nil or #savename == 0) then
    savename = "map-" .. map_id
  end
  local prefix = params.prefix .. savename .. "/"
  local data_dir = "d-" .. unique_id
  local data_prefix = prefix .. data_dir .. "/"
  game.print("Mapshot '" .. prefix .. "' ...")
  log("Mapshot target " .. prefix)
  log("Mapshot data target " .. data_prefix)
  log("Mapshot unique id " .. unique_id)

  local surface_infos = {}
  log("Request surface(s): " .. params.surface)
  for _, surface in pairs(game.surfaces) do
    local include_surface = should_render_surface(params, surface.name)
    log("Available surface: idx=" .. surface.index .. " name=" .. surface.name .. " include=" .. tostring(include_surface))
    if (include_surface) then
      local si = gen_surface_info(params, surface)
      if si then
        table.insert(surface_infos, si)
      end
    end
  end

  -- collect game version and active mod versions
  local game_version = game.active_mods["base"]
  local active_mods = {}
  for name, version in pairs(game.active_mods) do
    if name ~= "base" then
      active_mods[name] = version
    end
  end

  -- Write metadata.
  game.write_file(data_prefix .. "mapshot.json", game.table_to_json({
    savename = params.savename,
    unique_id = unique_id,
    map_id = map_id,
    tick = game.tick,
    ticks_played = game.ticks_played,
    seed = game.default_map_gen_settings.seed,
    map_exchange = game.get_map_exchange_string(),
    surfaces = surface_infos,
    game_version = game_version,
    active_mods = active_mods
  }))

  -- Create the serving html.
  for fname, contentfunc in pairs(generated.files) do
    local content = contentfunc()
    if (fname == "index.html") then
      local config = {
        path = data_dir,
      }
      content = string.gsub(content, "__MAPSHOT_CONFIG_TOKEN__", game.table_to_json(config))
    end
    local r = game.write_file(prefix .. fname, content)
  end

  -- Generate all the tiles.
  for _, surface_info in ipairs(surface_infos) do
    for render_zoom = surface_info.zoom_min, surface_info.zoom_max do
      local tile_size = surface_info.tile_size / math.pow(2, render_zoom)
      local layer_prefix = data_prefix .. surface_info.file_prefix .. render_zoom .. "/"
      gen_layer(params, tile_size, surface_info.render_size, surface_info.world_min, surface_info.world_max, layer_prefix, game.surfaces[surface_info.surface_idx])
    end
  end

  game.print("Mapshot: all screenshots started, might take a while to render; location: " .. data_prefix)
  log("Mapshot: all screenshots started, might take a while to render; location: " .. data_prefix)

  return data_prefix
end

-- Check if a surface should be rendered.
function should_render_surface(params, surface_name)
  if(params.surface == all_surfaces) then
    return true
  end

  -- Strip begin/end spaces from the input.
  surface_name = string.match(surface_name, "^%s*(.*)%s*$")

  -- Split params.surface into a list of surface names.
  local fields = {}
  string.gsub(params.surface, "([^,]+)", function(c) fields[#fields + 1] = c end)

  -- Check if the list contains the requested surface name.
  for _, f in ipairs(fields) do
    if (string.match(f, "^%s*(.*)%s*$") == surface_name) then
      return true
    end
  end

  return false
end

function gen_surface_info(params, surface)
  -- Determine map min & max world coordinates based on existing chunks.
  -- When requested to match only entities, fallback using all chunks
  -- if no entities are found at all.
  local try_ent_only = params.area == "entities"
  local world_min = { x = 2^30, y = 2^30 }
  local world_max = { x = -2^30, y = -2^30 }
  local chunk_count = 0
  local ent_world_min = { x = 2^30, y = 2^30 }
  local ent_world_max = { x = -2^30, y = -2^30 }
  local ent_chunk_count = 0
  for chunk in surface.get_chunks() do
    local in_all = surface.is_chunk_generated(chunk)
    if in_all then
      world_min.x = math.min(world_min.x, chunk.area.left_top.x)
      world_min.y = math.min(world_min.y, chunk.area.left_top.y)
      world_max.x = math.max(world_max.x, chunk.area.right_bottom.x)
      world_max.y = math.max(world_max.y, chunk.area.right_bottom.y)
      chunk_count = chunk_count + 1

      if try_ent_only then
        in_ent = surface.count_entities_filtered({ area = chunk.area, limit = 1, type = entities.includes}) > 0
        if in_ent then
          ent_world_min.x = math.min(ent_world_min.x, chunk.area.left_top.x)
          ent_world_min.y = math.min(ent_world_min.y, chunk.area.left_top.y)
          ent_world_max.x = math.max(ent_world_max.x, chunk.area.right_bottom.x)
          ent_world_max.y = math.max(ent_world_max.y, chunk.area.right_bottom.y)
          ent_chunk_count = ent_chunk_count + 1
        end
      end
    end
  end
  if try_ent_only and ent_chunk_count > 0 then
    world_min = ent_world_min
    world_max = ent_world_max
    chunk_count = ent_chunk_count
  end
  if chunk_count == 0 then
    log("no matching chunk")
    game.print("No matching chunk")
    return
  end
  game.print("Map: (" .. world_min.x .. ", " .. world_min.y .. ")-(" .. world_max.x .. ", " .. world_max.y .. ")")
  local area = {
    left_top = {world_min.x, world_min.y},
    right_bottom = {world_max.x, world_max.y},
  }

  -- Range of tiles to render, in power of 2.
  local tile_range_min = math.log(params.tilemin, 2)
  local tile_range_max = math.log(params.tilemax, 2)

  -- Size of a tile, in pixels.
  local render_size = params.resolution

  -- Find train stations
  local stations = {}
  for _, ent in ipairs(surface.find_entities_filtered({area=area, type="train-stop"})) do
    table.insert(stations, {
      backer_name = ent.backer_name,
      bounding_box = ent.bounding_box,
    })
  end

  -- Find all chart tags - aka, map labels.
  local tags = {}
  for _, force in pairs(game.forces) do
    for _, tag in ipairs(force.find_chart_tags(surface, area)) do
      table.insert(tags, {
        force_name = force.name,
        force_index = force.index,
        icon = tag.icon,
        tag_number = tag.tag_number,
        position = tag.position,
        text = tag.text,
      })
    end
  end

  local players = {}
  for _, player in pairs(game.players) do
    -- Make sure the player is on the current surface.
    if player.surface == surface then
      table.insert(players, {
        name = player.name,
        color = player.color,
        position = player.position,
      })
    end
  end

  return {
    surface_name = surface.name,
    surface_idx = surface.index,
    file_prefix = "s" .. surface.index .. "zoom_",
    tile_size = math.pow(2, tile_range_max),
    render_size = render_size,
    world_min = world_min,
    world_max = world_max,
    zoom_min = 0,
    zoom_max = tile_range_max - tile_range_min,
    players = players,
    stations = stations,
    tags = tags,
  }
end

function gen_layer(params, tile_size, render_size, world_min, world_max, data_prefix, surface)
  local tile_min = { x = math.floor(world_min.x / tile_size), y = math.floor(world_min.y / tile_size) }
  local tile_max = { x = math.floor(world_max.x / tile_size), y = math.floor(world_max.y / tile_size) }

  local msg =  "Tile size " .. tile_size .. ": " .. (tile_max.x - tile_min.x + 1) * (tile_max.y - tile_min.y + 1) .. " tiles to generate"
  game.print(msg)
  log(msg)

  for tile_y = tile_min.y, tile_max.y do
    for tile_x = tile_min.x, tile_max.x do
      local top_left = { x = tile_x * tile_size, y = tile_y * tile_size }
      local bottom_right = { x = top_left.x + tile_size, y = top_left.y + tile_size }
      local has_entities = surface.count_entities_filtered({ area = {top_left, bottom_right}, limit = 1, type = entities.includes}) > 0
      local quality_to_use = has_entities and params.jpgquality or math.min(params.minjpgquality, params.jpgquality)
      if quality_to_use > 0 then
        game.take_screenshot{
          surface = surface,
          position = {
            x = top_left.x + tile_size / 2,
            y = top_left.y + tile_size / 2,
          },
          resolution = {render_size, render_size},
          zoom = factorio_zoom(render_size, tile_size),
          path = data_prefix .. "tile_" .. tile_x .. "_" .. tile_y .. ".jpg",
          show_gui = false,
          show_entity_info = true,
          quality = quality_to_use,
          daytime = 0,
          water_tick = 0,
        }
      end
    end
  end
end

-- Create a unique ID of the generated mapshot.
function gen_unique_id()
  local data = generated.version_hash .. " " .. tostring(game.tick) .. " " .. game.get_map_exchange_string()
  -- sha256 produces 64 digits. We're not looking for crypto secure hashing, and instead
  -- just a short unique string - so pick up a subset.
  local idx = 10
  local len = 8
  local h = string.sub(hash.hash256(data), idx, idx + len - 1)
  log("Unique ID: " .. h)
  return h
end

-- Create a unique ID for the game being played.
function gen_map_id()
  -- sha256 produces 64 digits. We're not looking for crypto secure hashing, and instead
  -- just a short unique string - so pick up a subset.
  local idx = 10
  local len = 8
  local h = string.sub(hash.hash256(game.get_map_exchange_string()), idx, idx + len - 1)
  log("Map ID: " .. h)
  return h
end

-- Detects if an on-startup screenshot is requested.
script.on_event(defines.events.on_tick, function(evt)
  log("onstartup check @" .. evt.tick)
  -- Needs to run only once, so unregister immediately.
  script.on_event(defines.events.on_tick, nil)

  -- Assume player index 1 during startup.
  local player = game.get_player(1)
  local params = build_params(player)

  if params.onstartup ~= "" then
    log("onstartup requested id=" .. params.onstartup)
    local data_prefix = mapshot(params)

    -- Ensure that screen shots are written before marking as done.
    game.set_wait_for_screenshots_to_finish()

    -- When set_wait_for_screenshots_to_finish was not used, the `done` file was
    -- be written before the screenshots, leading to killing Factorio too early.
    -- On Linux, using signal Interrupt helped a lot, but that did not guarantee
    -- it - and it is not available on Windows. Writing the `done` marker on the
    -- next tick seemed enough to guarantee ordering. Now
    -- set_wait_for_screenshots_to_finish is used, this is likely unnecessary -
    -- but before removing it, more testing is needed.
    script.on_event(defines.events.on_tick, function(evt)
      log("marking as done @" .. evt.tick)
      script.on_event(defines.events.on_tick, nil)
      game.write_file("mapshot-done-" .. params.onstartup, data_prefix)
    end)
  end
end)

-- Register the command.
-- It seems that on_init+on_load sometime don't trigger (neither of them) when
-- doing weird things with --mod-directory and list of active mods.
commands.add_command("mapshot", "screenshot the whole map", function(evt)
  local player = game.get_player(evt.player_index)
  local params = build_params(player)
  if evt.parameter ~= nil and #evt.parameter > 0 then
    params.savename = evt.parameter
  end
  mapshot(params)
end)
