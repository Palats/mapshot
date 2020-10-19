# Development

Most instructions here assume a Linux host. The working directory is the root of
the checkout.

## Initial setup

You need to have working installs of git, Go, and NPM.

```
git checkout https://github.com/Palats/mapshot.git
cd mapshot
npm --prefix frontend install
```

## Development cycle

Keep an automatic rebuild of the frontend in background:
```
npm --prefix frontend run watch
```
This will automatically rebuild `frontend/dist/*` files on change.

Then, run the CLI in `dev` mode:
```
go run mapshot.go dev
```

This will:
 - Start Factorio with a customized list of mod, including the mapshot mod.
 - The mapshot mod will link directly the files in the checkout - so
   modifications of the lua files will be reflected when loading a savegame in
   Factorio.
 - A HTTP server will run in backgrdound to access generated content.
 - The UI frontend will use the file watched by the `npm` command - so any
   changes to the frontend code will be reflected on a page reload.

The main limitation of the `dev` mode is that files is that the fiels embedded in the CLI binary will not be automatically updated. In practice, the only consequence is that generated `index.html` and companions files will not be up to date. However, the HTTP server in the CLI do not use them anyway.

The frontend can load arbitrary mapshots by adding `?path=mapshot/<name>` query parameters.

The files in the `mod` directory of the repository can be used directly by
Factorio. This allows to a quick edit/test cycle. That directory can be linked
from your Factorio `mods/` directory under the name `mapshot`.

This will run Factorio with customized list of mods, including the mapshot mod - using links directly to the repository, so changes will be visible in Factorio after reload a save. Don't forget that generated content will not be automatically updated.

## Regenerating files

To have a clean build, multiple files need to be generated.

If source in `frontend/` is changed:
```
npm --prefix frontend run build
```

This builds the compiled version of the HTML/Javascript frontend bits in
`frontend/dist/*`. Those files are _not_ committed to the repository, as per
frontend common practices.

If source in `mod/` or `frontend/` is changed, after running `npm`:
```
go generate ./...
```

It generates:
- `mod/generated.lua`, used directly in the Factorio mod. It contains data
   necessary for the mod to be able to generate the website for each render; it
   comes from the `frontend/dist/*` files.
- `embed/generated.go`, used in the CLI. It contains all the mod files in a
   format accessible from Go.

Those files _are_ committed to the repository, as per `go generate` common
practices.

## Releasing

* Update `changelog.txt`
* Update version in: `changelog.txt` (add date), `mod/info.json`, `frontend/package.json`
* Regenerate files: `npm --prefix frontend run build && go generate ./...`
* Test build release: `./build.sh`
* Commit and push
* Create release in Github
* Update Factorio mods portal (new zip, update doc)