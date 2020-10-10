# Development

Most instructions here assume a Linux host.

## Initial setup

You need to have working installs of git, Go, and NPM.

```
git checkout https://github.com/Palats/mapshot.git
cd mapshot
npm --prefix frontend install
```

## Regenerating files

Multiple files are automatically generated and thus need to be updated.

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

## Working on the mod

The files in the `mod` directory of the repository can be used directly by
Factorio. This allows to a quick edit/test cycle. That directory can be linked
from your Factorio `mods/` directory under the name `mapshot`.

The CLI also contains a more convenient alternative:
```
go run mapshot.go dev
```

This will run Factorio with customized list of mods, including the mapshot mod - using links directly to the repository, so changes will be visible in Factorio after reload a save. Don't forget that generated content will not be automatically updated.

## Working on the CLI

To run it from a checkout of the repository:
```
go run mapshot.go <parameters...>
```

By default, it will show the help, including all the available subcommands. Note
that it will run with the currently generated content, which might not be up to
date - see "Regenerating" section.

## Working on the UI

To automatically rebuild `frontend/dist/*` files on change:
```
npm --prefix frontend run watch
```

The frontend can load arbitrary mapshots by adding `?path=mapshot/<name>` query parameters.

## Releasing

* Update `changelog.txt`
* Update version in: `changelog.txt` (add date), `mod/info.json`, `frontend/package.json`
* Regenerate files: `npm --prefix frontend run build && go generate ./...`
* Test build release: `./build.sh`
* Commit and push
* Create release in Github
* Update Factorio mods portal (new zip, update doc)