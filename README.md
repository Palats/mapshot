# Mapshot for Factorio

*Mapshot* generates zoomable screenshots of Factorio maps - **[example](https://palats.github.io/mapshot-example/)**.

They can be created in 2 ways:

* Through a regular Factorio mod, providing an extra command to create a mapshot.
* Through a CLI tool generating a mapshot of any saved game - without having to activate mods on your game. Factorio is used for rendering.

The generated zoomable screenshots can be explored through a web browser, using the CLI tool to serve them. As those mapshots are exported as static files (html, javascript, jpg), they can also be served through any HTTP server - see below.

Some simple layers are generated - currently it is possible to show train stations and map labels (chart tags).

***Warning: Generation can take quite a while. Factorio UI will appear frozen during that time; this is normal.***

See https://github.com/Palats/mapshot for more details.

## Installing

The Factorio mod can be installed like any other mod from the Factorio UI.

The optional CLI is used for serving generated mapshots and generating mapshots from outside the game. The standalone binaries can be downloaded from https://github.com/Palats/mapshot/releases; then:

 * Linux: Mark as executable if needed and run - this is a standard command line tool.
 * Windows: For convenience, if the `.exe` file is launched directly from Explorer, it will automatically start the serving mode. Otherwise, you need a way to give the tool parameters - either by launching it from the `cmd` console, or by creating a shortcut (with extra parameters in the properties).
 * MacOS: A binary is provided ("darwin" version), but is completely untested as I have no access to a MacOS system.


## Creating a mapshot

### In Factorio

In the Factorio [console](https://wiki.factorio.com/Console), run:
```
/mapshot <name>
```

It will create a website in the Factorio [script output directory](https://wiki.factorio.com/Application_directory#User_data_directory), in a subdirectory called `mapshot/<name>/`. If `<name>` is not specified, a name based on the seed and current tick will be generated (modding API does not gives access to savename, hence no good default naming).

### With the CLI

To generate a mapshot:

```
./mapshot render <savename>
```
where `savename` is the name of the save you want to render. It will not modify the file - specifically, despite mod usage, it will not impact achievements for example. This will run Factorio to generate the mapshot - let it run, it will shut it down when finished. As with the regular mod, the output will be somewhere in the `script-output` directory.

If your Factorio data dir or binary location are not detected automatically, you can specify them with `--factorio_datadir` and `--factorio_binary`. You can also override the rendering parameters - see CLI help for the specific flag names.

Steam version of Factorio is not supported for now. If you have only a Steam version, you can still get a standalone version on factorio.com by linking your Steam account.

### Parameters

You can tune parameters such as many layers to generate, their resolution and a few more details. Those parameters are:

* ... in Factorio per-player mod settings interface (in the menu, `settings/Mod settings/Per player`).
* ... command line arguments for the CLI.

Parameters:

* _Area_ (`area`) : What to include in the mapshot. Options:
  * `entities` [default]: Include all chunks which contain at least one entity of some interest. This should capture the base in practice.
  * `all`: All chunks.
* _Smallest tile size_ (`tilemin`) : Indicates the number of in-game units the most detailed layer should contain per generated tile. For example, if it is set to 256 while the "Tile Resolution" is 1024, it means that the most detailed layer will use 4 pixels (=1024/256) per in-game tile. Many assets in Factorio seem to allow for up to 64 pixels per game tile - so, to have the maximum resolution, you will want to have "Smallest tile size" set to 16 (=1024/64) - careful, that is slow.
* _Largest tile size_ (`tilemax`) : Number of in-game units per generated tile for the least detailed layer. See `tilemin` for more details. Mapshot will generates all layers from `tilemax` to `tilemin` (included).
* _Prefix to add to all generated filenames._ (`prefix`) : Mapshot will prefix all files it creates with that value. Factorio mods only allow writing within `script-output` subdirectory of Factorio data dir; the prefix is relative to that directory.
* _Pixel size for generated tiles._ (`resolution`) : Size in pixels for the generated images. There is not a lot of reasons to change this value - if you want more or less details, change `tilemin`.
* _Pixel size for generated tiles._ (`jpgquality`) : Compression quality for the generated image.

*Warning: the generation time & disk usage increases very quickly. At maximum resolution, it will take forever to generate and use up several gigabytes of space.*


## Serving the generated content

The CLI can be used to serve the mapshots:

```
./mapshot serve
```

By default, it serves on port 8080 - thus accessible at http://localhost:8080 if it is running on your local machine. It serves all the mapshots available in the `script-output` directory of Factorio. It provides a very basic list of available mapshots and refreshes this list every few seconds.

The generated content is made of static files. This means you can also serve the content through any HTTP server (e.g., `python3 -m http.server 8080` from the `script-output` directory) or your favorite web file hosting.


## Development

See [DEVELOPMENT.md](#DEVELOPMENT.md) in the repository.