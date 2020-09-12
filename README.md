# `mapshot` for Factorio
Factorio mod to generate a screenshot of the whole map. The screenshot is created as a website, allowing to zoom-in and navigate the map.

## Usage

### Generating a map screenshot
In the Factorio [console](https://wiki.factorio.com/Console), run:
```
/mapshot <name>
```

***Warning: This can take quite a while. Factorio UI will appear frozen during that time; this is normal.***

It will create a website in the Factorio [script output directory](https://wiki.factorio.com/Application_directory#User_data_directory), in a subdirectory called `mapshot/<name>/`.

If `<name>` is not specified, a name based on the seed and current tick will be generated (modding API does not gives access to savename, hence no good default naming).

### Serving the generated content
The mod itself does not provide any facility to serve the content. You must make the content available through a standard http server. All the content is static, so any http file server can do the trick. For example, if you go in `script-output/mapshot/<name>`, you can run a simple server using Python:
```
python3 -m http.server 8080
```

You can also upload the result on your favorite web file hosting.

### Generation parameters

You can tune parameters such as many layers to generate, their resolution and a few more details. Those parameters are accessible in Factorio per-player mod settings interface (in the menu, `settings/Mod settings/Per player`).

The most important parameter is the "Smallest tile size". It indicates the number of in-game units the most detailed layer should contain. For example, if it is set to 256 while the "Tile Resolution" is 1024, it means that the most detailed layer will use 4 pixels (=1024/256) per in-game tile. Afaik, assets in Factorio allow for up to 64 pixels per game tile - so, to have the maximum resolution, you will want to have "Smallest tile size" set to 16 (=1024/64).

***Warning: the generation time & disk usage increases very quickly. At maximum resolution, it will take forever to generate and use up several gigabytes of space.***


## Development

* Check out the repository.
* Link your checkout from the Factorio `mods` directory under the name `mapshot`.
* To avoid having to regenerate Lua files when modifying the html part of the plugin, link `viewer.html` from `script-output`. Load `viewer.html` in your browser and add `?path=mapshot/<name>` to look directly at one of the generated map.

Files in `embed/` and `generated.lua` are automatically generated from other files of the repository; to regenerate them:
```
go run genembed.go
```

### Releasing

* Update changelog
* Update version in: `changelog.txt` (incl. date), `info.json`
* Regenerate files: `go run genembed.go`
* Commit and push
* Build CLI: `go build mapshot.go`
* Build mod: `./mapshot package`
* Create release in Github
* Update Factorio mods portal (new zip, update doc)