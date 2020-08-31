script.on_load(function()
  commands.add_command("mapshot", "screenshot the whole map", mapshot)
end)

function mapshot(evt)
    -- Determine map min & max world coordinates based on existing chunks.
    local world_min = { x = 2^30, y = 2^30 }
    local world_max = { x = -2^30, y = -2^30 }
    for chunk in game.surfaces["nauvis"].get_chunks() do
      world_min.x = math.min(world_min.x, chunk.area.left_top.x)
      world_min.y = math.min(world_min.y, chunk.area.left_top.y)
      world_max.x = math.max(world_max.x, chunk.area.right_bottom.x)
      world_max.y = math.max(world_max.y, chunk.area.right_bottom.y)
    end
    game.print("Map: (" .. world_min.x .. ", " .. world_min.y .. ")-(" .. world_max.x .. ", " .. world_max.y .. ")")

    -- Size of a tile, in world coordinates.
    local tile_size = 128
    -- Size of a tile, in pixels
    local res_size = 1024

    -- Zoom. We want to have res_size pixels represent tile_size world unit.
    -- A zoom of 1.0 means that 32 pixels represent 1 world unit. A zoom of 2.0 means 64 pixels per world unit.
    local zoom = res_size / 32 / tile_size

    -- Modulo is always positive. -2 % 5 == 3.
    -- math.fmod(-2, 5) == -2
    -- math.floor(-2 / 5) == -1

    local tile_min = { x = math.floor(world_min.x / tile_size), y = math.floor(world_min.y / tile_size) }
    local tile_max = { x = math.floor(world_max.x / tile_size), y = math.floor(world_max.y / tile_size) }

    game.print((tile_max.x - tile_min.x + 1) * (tile_max.y - tile_min.y + 1) .. " tiles to generate")

    local prefix = "mapshot/"

    -- Write metadata.
    game.write_file(prefix .. "map.json", game.table_to_json({
      tile_size = tile_size,
      res_size = res_size,
      world_min = world_min,
      world_max = world_max,
    }))

    -- Generate all tiles
    for tile_y = tile_min.y, tile_max.y + 1 do
      for tile_x = tile_min.x, tile_max.x + 1 do
        -- game.print("x=" .. tile_x .. " y=" .. tile_y)
        local top_left = { x = tile_x * tile_size, y = tile_y * tile_size }
        game.take_screenshot{
          position = {
            x = top_left.x + tile_size / 2,
            y = top_left.y + tile_size / 2,
          },
          resolution = {res_size, res_size},
          zoom = zoom,
          path = prefix .. "tile_" .. tile_x .. "_" .. tile_y .. ".jpg",
          show_gui = false,
          show_entity_info = true,
          quality = 75,
          daytime = 0,
          water_tick = 0,
        }
      end
    end
end