local generated = require("generated")
local overrides = require("overrides")
local entities = require("entities")
local hash = require("hash")

local factorio_min_zoom = 0.031250

-- Read all settings and update the params var, incl. overrides.
function build_params(player)
  local params = {}
  -- settings.player[xxx] does contain the value at the beginning of the game,
  -- while get_player_settings contains the current value.
  local s = settings.get_player_settings(player)
  for k, v in pairs(s) do
    params[k] = v.value
  end

  for k,v in pairs(game.json_to_table(overrides)) do
    params[k] = v
  end

  if (string.sub(params.prefix, -1) ~= "/") then
    params.prefix = params.prefix .. "/"
  end

  params.tilemin = factorio_fit_zoom(params.resolution, params.tilemin, "tilemin", player)
  params.tilemax = factorio_fit_zoom(params.resolution, params.tilemax, "tilemax", player)

  return params
end

-- Calculate the zoom value for Factorio take_screenshot function
function factorio_zoom(render_size, tile_size)
  -- We want to have render_size pixels represent tile_size world unit.
  -- A zoom of 1.0 means that 32 pixels represent 1 world unit. A zoom of 2.0 means 64 pixels per world unit.
  return render_size / 32 / tile_size
end

function factorio_fit_zoom(render_size, tile_size, name, player)
  if (factorio_zoom(render_size, tile_size) < factorio_min_zoom) then
    local old = tile_size
    tile_size = render_size / 32 / factorio_min_zoom
    local msg = "Parameter " .. name .. " changed from " .. old .. " to " .. tile_size .. " to fit within Factorio minimal zoom of " .. factorio_min_zoom
    player.print(msg)
    log(msg)
  end
  return tile_size
end

-- Generate a full map screenshot.
function mapshot(player, params)
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
  player.print("Mapshot '" .. prefix .. "' ...")
  log("Mapshot target " .. prefix)
  log("Mapshot data target " .. data_prefix)
  log("Mapshot unique id " .. unique_id)

  local surface = game.surfaces["nauvis"]

  -- Determine map min & max world coordinates based on existing chunks.
  local world_min = { x = 2^30, y = 2^30 }
  local world_max = { x = -2^30, y = -2^30 }
  local chunk_count = 0
  for chunk in surface.get_chunks() do
    local c = surface.is_chunk_generated(chunk)
    if params.area == "entities" then
      c = c and surface.count_entities_filtered({ area = chunk.area, limit = 1, type = entities.includes}) > 0
    end
    if c then
      world_min.x = math.min(world_min.x, chunk.area.left_top.x)
      world_min.y = math.min(world_min.y, chunk.area.left_top.y)
      world_max.x = math.max(world_max.x, chunk.area.right_bottom.x)
      world_max.y = math.max(world_max.y, chunk.area.right_bottom.y)
      chunk_count = chunk_count + 1
    end
  end
  if chunk_count == 0 then
    log("no matching chunk")
    player.print("No matching chunk")
    return
  end
  player.print("Map: (" .. world_min.x .. ", " .. world_min.y .. ")-(" .. world_max.x .. ", " .. world_max.y .. ")")
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
  for _, ent in ipairs(surface.find_entities_filtered({area=area, name="train-stop"})) do
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
    table.insert(players, {
      name = player.name,
      color = player.color,
      position = player.position,
    })
  end

  -- Write metadata.
  game.write_file(data_prefix .. "mapshot.json", game.table_to_json({
    savename = params.savename,
    unique_id = unique_id,
    map_id = map_id,
    tick = game.tick,
    ticks_played = game.ticks_played,
    tile_size = math.pow(2, tile_range_max),
    render_size = render_size,
    world_min = world_min,
    world_max = world_max,
    zoom_min = 0,
    zoom_max = tile_range_max - tile_range_min,
    seed = game.default_map_gen_settings.seed,
    map_exchange = game.get_map_exchange_string(),
    players = players,
    stations = stations,
    tags = tags,
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
  for tile_range = tile_range_max, tile_range_min, -1 do
    local tile_size = math.pow(2, tile_range)
    local render_zoom = tile_range_max - tile_range
    gen_layer(player, params, tile_size, render_size, world_min, world_max, data_prefix .. "zoom_" .. render_zoom .. "/")
  end

  player.print("Mapshot done at " .. data_prefix)
  log("Mapshot done at " .. data_prefix)

  return data_prefix
end

function gen_layer(player, params, tile_size, render_size, world_min, world_max, data_prefix)
  local tile_min = { x = math.floor(world_min.x / tile_size), y = math.floor(world_min.y / tile_size) }
  local tile_max = { x = math.floor(world_max.x / tile_size), y = math.floor(world_max.y / tile_size) }

  local msg =  "Tile size " .. tile_size .. ": " .. (tile_max.x - tile_min.x + 1) * (tile_max.y - tile_min.y + 1) .. " tiles to generate"
  player.print(msg)
  log(msg)

  for tile_y = tile_min.y, tile_max.y do
    for tile_x = tile_min.x, tile_max.x do
      local top_left = { x = tile_x * tile_size, y = tile_y * tile_size }
      game.take_screenshot{
        position = {
          x = top_left.x + tile_size / 2,
          y = top_left.y + tile_size / 2,
        },
        resolution = {render_size, render_size},
        zoom = factorio_zoom(render_size, tile_size),
        path = data_prefix .. "tile_" .. tile_x .. "_" .. tile_y .. ".jpg",
        show_gui = false,
        show_entity_info = true,
        quality = params.jpgquality,
        daytime = 0,
        water_tick = 0,
      }
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
    local data_prefix = mapshot(player, params)

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
  mapshot(player, params)
end)