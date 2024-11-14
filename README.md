# Mapshot for Factorio

*Mapshot* generates zoomable screenshots of Factorio maps - **[example](https://mapshot.palats.xyz/)**.

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

An externally maintained package for Arch [is also available](https://aur.archlinux.org/packages/mapshot), thanks to [Sharparam](https://github.com/Sharparam).

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

This commands takes the list of mods to load from the current active list in Factorio config - not the list of mods which were used when the save was created. This can lead to inconsistent renders; see https://github.com/Palats/mapshot/issues/20.

If your Factorio data dir or binary location are not detected automatically, you can specify them with `--factorio_datadir` and `--factorio_binary`. You can also override the rendering parameters - see CLI help for the specific flag names.

Steam version of Factorio is not supported for now - see https://github.com/Palats/mapshot/issues/21 for more details. If you have only a Steam version, you can still get a standalone version on factorio.com by linking your Steam account.

> [!WARNING]
> Headless version of Factorio is not supported as it lacks the ability to render any image.
> Most common symptom is Factorio failing with `Option ‘disable-audio’ does not exist`.

### Parameters

You can tune parameters such as many layers to generate, their resolution and a few more details. Those parameters are:

* ... in Factorio per-player mod settings interface (in the menu, `settings/Mod settings/Per player`).
* ... command line arguments for the CLI.

Parameters:

* _Area_ (`area`) : What to include in the mapshot. Options:
  * `entities` [default]: Include all chunks which contain at least one entity of some interest. This should capture the base in practice. When no entities are found, all chunks are rendered.
  * `all`: All chunks.
* _Smallest tile size_ (`tilemin`) : Indicates the number of in-game units the most detailed layer should contain per generated tile. For example, if it is set to 256 while the "Tile Resolution" is 1024, it means that the most detailed layer will use 4 pixels (=1024/256) per in-game tile. Many assets in Factorio seem to allow for up to 64 pixels per game tile - so, to have the maximum resolution, you will want to have "Smallest tile size" set to 16 (=1024/64) - careful, that is slow.
* _Largest tile size_ (`tilemax`) : Number of in-game units per generated tile for the least detailed layer. See `tilemin` for more details. Mapshot will generates all layers from `tilemax` to `tilemin` (included).
* _Prefix to add to all generated filenames._ (`prefix`) : Mapshot will prefix all files it creates with that value. Factorio mods only allow writing within `script-output` subdirectory of Factorio data dir; the prefix is relative to that directory.
* _Pixel size for generated tiles._ (`resolution`) : Size in pixels for the generated images. There is not a lot of reasons to change this value - if you want more or less details, change `tilemin`.
* _Pixel size for generated tiles._ (`jpgquality`) : Compression quality for the generated image.
* _Pixel size for generated tiles._ (`minjpgquality`) : Compression quality for the generated image when no player entities are present. If set to 0, do not render a tile at all; instead, the map rendering will fallback to a lower zoom level as needed.
* _Surface name._ (`surface`) : Restrict which game surface to generate, defaulting to `_all_`, which generate shots of all surfaces.

*Warning: the generation time & disk usage increases very quickly. At maximum resolution, it will take forever to generate and use up several gigabytes of space.*

### Headless server

NOTE: This is hacky and you need some familiarity with Linux; also, you are mostly on your own.

Mapshot requires a running Factorio with UI to do the rendering - this is a constraint of Factorio itself. On Linux, this means a X server must be available. On a Linux headless server, it is still possible to do renders using [Xvfb](https://en.wikipedia.org/wiki/Xvfb).

On Ubuntu, Xvfb can be installed through `apt-get install xvfb`. Once you have it installed, you can create a mapshot by running the command through `xvfb-run`; for example:
```
xvfb-run ./mapshot render <savename>
```

It can be a bit fiddly with OpenGL; a few tips:

* Make sure you have a recent version of Xvfb / distro. For example, on Ubuntu 18.04 there are issues with OpenGL, while it works fine on Ubuntu 20.04.
* https://github.com/Palats/mapshot/issues/8 has a suggestion using virtualgl.
* https://github.com/Palats/mapshot/issues/16 has other suggestions.
* https://github.com/Palats/mapshot/issues/53 has some discussions and examples for Docker / Dockerfile .
* For Factorio 1.1.36 (and probably later, until fixed), https://github.com/Palats/mapshot/issues/16#issuecomment-883306221 has a suggested solution.

## Serving the maps

The CLI can be used to serve the mapshots:

```
./mapshot serve
```

By default, it serves on port 8080 - thus accessible at http://localhost:8080 if it is running on your local machine. It serves all the mapshots available in the `script-output` directory of Factorio. Directory can be overriden using flag `--factorio_scriptoutput`. It provides a very basic list of available mapshots and refreshes this list every few seconds. (Note: it uses frontend code built into the binary. It ignores the frontend files such as `index.html` and Javascript files present next to the mapshots.)

The generated content has static frontend code generated next to the images. This means you can also serve the content through any HTTP server (e.g., `python3 -m http.server 8080` from the `script-output` directory) or your favorite web file hosting.

The viewer has the following URL query parameters:

* `path`: string, URL of the mapshot to display.
* `x`, `y`: float, center position in Factorio coordinates.
* `z` : float, zoom level.
* `lt`, `lg`, `ld`: "0"|"1", show/hide various layers (train stations, tags, debug).


## Generated content

### Directory hierarchy

All content is generated in the Mapshot output directory. This directory is `script-output/<prefix>`, where:

* `script-output` is the default Factorio directory where mods can write.
* `<prefix>` is a subdirectory where Mapshot can write. By default this is `mapshot/`.

Within that directory, a directory will be created per save:

* When using Factorio command `/mapshot <savename>`, the name will be `<savename>/`.
* When using Factorio command `/mapshot`, a savename will be generated, stable across invocation on the same game. This is based on map generation parameters, in the form `map-<hash>`.
* When using CLI `mapshot render <savename>`, the name will be `<savename>/`.

Within a given `<savename>` directory, one subdirectory will be created everytime a mapshot is made. It is of the form `d-<hash>`, where the hash is computed based on many input to try to be as unique as possible. Those directories contain some more internal directories to organize the raw data.

### Files

Currently no files are created in the Mapshot output directory itself.

In a `<savename>` directory, html and javascript files are created. It points to latest mapshot generated in that `<savename>` directory. Currently, accessing older mapshots require fiddling with `?path=xxx` URL query parameter.

In a given mapshot directory (of the form `d-<hash>`), a `mapshot.json` file describes that specific render.

### Caching

Generated `html` files are not meant to be cached, as they are potentially updated on each render. Javascript files can be cached as their name will change as needed. The `thumbnail.png` is used only as a favicon - while it might change in the future, it is not critical. Anything under a specific mapshot directory (`d-<hash>`) is immutable and can be cached indefinitely.

In practice, if adding a caching layer in front of `./mapshot serve`, everything can be cached as most of the content URLs contain hashes. Exceptions:

* `/` (not subpaths) is the mapshot view HTML and listing UI. It changes rarely - on every new release of the mod. Caching is likely fine for hours. Note: this is content from the `serve` command, not the html/js files generated in `script-output`.
* `/shots.json` is the list of available mapshots. It changes content in place everytime a new one mapshot is created. Caching should probably be short term to allow to see new content.
* `/latest/*` is information to link to the latest version of a given save. It can change when a new mapshot is created.
* `/thumbnail.png` is used as a favicon. It might change in place in the future, but can be cached heavily as it is non-critical.

### Example

Visually, that gives something like that:
```
script-output/mapshot/  <--- output directory
  savename1/
    index.html
    main-1c3f7217.js
    thumbnail.png
    d-5bd8e540/
      mapshot.json
      s1zoom_0/ ...
      s1zoom_1/ ...
      ...
    d-a309ff22/
      ...
    ...
  savename2/
    ...
  map-ad765988/
    ...
  ...
```

## Development

See [DEVELOPMENT.md](https://github.com/Palats/mapshot/blob/master/DEVELOPMENT.md) in the repository.