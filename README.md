# Mapshot for Factorio

*Mapshot* generates zoomable screenshots of Factorio maps. It can be used in 2 ways:

* As a regular Factorio mod, providing an extra command to create a screenshot.
* As a tool (CLI) generating a screenshot of any saved game - without having to activate mods on your game. Factorio is still used for rendering.

The zoomable screenshots are exported as static files (html, javascript, jpg). They can be served through any HTTP server - see below.

***Warning: Generation can take quite a while. Factorio UI will appear frozen during that time; this is normal.***

See https://github.com/Palats/mapshot for more details.

## Usage: Factorio mod

In the Factorio [console](https://wiki.factorio.com/Console), run:
```
/mapshot <name>
```

It will create a website in the Factorio [script output directory](https://wiki.factorio.com/Application_directory#User_data_directory), in a subdirectory called `mapshot/<name>/`. If `<name>` is not specified, a name based on the seed and current tick will be generated (modding API does not gives access to savename, hence no good default naming).

## Usage: CLI

You can download the latest binary in https://github.com/Palats/mapshot/releases . The CLI is a standalone binary, currently only for linux. To generate a screenshot:

```
./mapshot render <savename>
```
where `savename` is the name of the save you want to render. It will not modify the file - specifically, despite mod usage, it will not impact achievements for example. This will run Factorio to generate the mapshot - let it run, it will shut it down when finished. As with the regular mod, the output will be somewhere in the `script-output` directory.

If your Factorio data dir or binary location are not detected automatically, you can specify them with `--factorio_datadir` and `--factorio_binary`. You can also override the rendering parameters - see CLI help for the specific flag names.

## Generation parameters

You can tune parameters such as many layers to generate, their resolution and a few more details. Those parameters are:

* ... in Factorio per-player mod settings interface (in the menu, `settings/Mod settings/Per player`).
* ... command line arguments for the CLI.

Parameters:

* _Smallest tile size_ (`tilemin`) : Indicates the number of in-game units the most detailed layer should contain per generated tile. For example, if it is set to 256 while the "Tile Resolution" is 1024, it means that the most detailed layer will use 4 pixels (=1024/256) per in-game tile. Many assets in Factorio seem to allow for up to 64 pixels per game tile - so, to have the maximum resolution, you will want to have "Smallest tile size" set to 16 (=1024/64) - careful, that is slow.
* _Largest tile size_ (`tilemax`) : Number of in-game units per generated tile for the least detailed layer. See `tilemin` for more details. Mapshot will generates all layers from `tilemax` to `tilemin` (included).
* _Prefix to add to all generated filenames._ (`prefix`) : Mapshot will prefix all files it creates with that value. Factorio mods only allow writing within `script-output` subdirectory of Factorio data dir; the prefix is relative to that directory.
* _Pixel size for generated tiles._ (`resolution`) : Size in pixels for the generated images. There is not a lot of reasons to change this value - if you want more or less details, change `tilemin`.
* _Pixel size for generated tiles._ (`jpgquality`) : Compression quality for the generated image.

*Warning: the generation time & disk usage increases very quickly. At maximum resolution, it will take forever to generate and use up several gigabytes of space.*


## Serving the generated content
The mod itself does not provide any facility to serve the content. You must make the content available through a standard http server. All the content is static, so any http file server can do the trick. For example, if you go in `script-output/mapshot/<name>`, you can run a simple server using Python:
```
python3 -m http.server 8080
```

You can also upload the result on your favorite web file hosting.

## Development

### Running as a live mod

The files in a checkout of the repository can be used directly by Factorio. This allows to a quick edit/test cycle. For that, simply your checkout from the Factorio `mods` directory under the name `mapshot`.

To avoid having to regenerate Lua files when modifying the html part of the plugin, link `viewer.html` from `script-output`. Load `viewer.html` in your browser and add `?path=mapshot/<name>` to look directly at one of the generated map.

Files in `embed/` and `generated.lua` are automatically generated from other files of the repository; to regenerate them:
```
go generate ./...
```

### The CLI

To run it from a checkout of the repository:
```
go generate ./... && go run mapshot.go <parameters...>
```

By default, it will show the help, including all the available subcommands.

### Releasing

* Update changelog
* Update version in: `changelog.txt` (incl. date), `info.json`
* Regenerate files: `go generate ./...`
* Run tests
* Commit and push
* Build CLI: `go build mapshot.go`
* Build mod: `./mapshot package`
* Create release in Github
* Update Factorio mods portal (new zip, update doc)