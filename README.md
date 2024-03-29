# csgo-centrifuge

Go API & CLI for downloading and extracting data from BSP files.
Can be used to get radar-overviews for all historic map versions of CS:GO.

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/saiko-tech/csgo-centrifuge/pkg?tab=doc)
[![Go Report](https://goreportcard.com/badge/github.com/saiko-tech/csgo-centrifuge?style=flat-square)](https://goreportcard.com/report/github.com/saiko-tech/csgo-centrifuge)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE.md)
[![Better Uptime Badge](https://betteruptime.com/status-badges/v1/monitor/mgn5.svg)](https://betteruptime.com/?utm_source=status_badge)

## API / Public Service

We provide a free public service to get any radar image and map-info file easily.

### Usage

#### URL Formats
- `https://radar-overviews.csgo.saiko.tech/<map>/<crc>/radar.dds`
- `https://radar-overviews.csgo.saiko.tech/<map>/<crc>/radar.png` - `radar.dds` converted to PNG (smaller, same quality)
- `https://radar-overviews.csgo.saiko.tech/<map>/<crc>/info.txt` - VDF format (Valve Data Format)
- `https://radar-overviews.csgo.saiko.tech/<map>/<crc>/info.json` - `info.txt` converted to JSON (easier to work with)
- `https://radar-overviews.csgo.saiko.tech/<map>/<crc>/nav.nav` - `<map>.nav`- see https://developer.valvesoftware.com/wiki/Nav_Mesh and https://developer.valvesoftware.com/wiki/NAV

ℹ️ See further down on [how to get the `<crc>` value](#how-to-get-the-map-crc-code).

#### Examples

- https://radar-overviews.csgo.saiko.tech/cs_agency/2230463619/radar.dds
- https://radar-overviews.csgo.saiko.tech/cs_agency/2230463619/radar.png
- https://radar-overviews.csgo.saiko.tech/cs_agency/2230463619/info.txt
- https://radar-overviews.csgo.saiko.tech/cs_agency/2230463619/info.json
- https://radar-overviews.csgo.saiko.tech/cs_agency/2230463619/nav.nav

### Limitations / Contact

The public service does not offer any uptime or compatibility guarantees.
It also doesn't allow listing/indexing of all map versions or access to the BSP files.

If you are interested in a business-grade service (and many more features!) please get in touch - you can find contact details on the [saiko.tech](https://saiko.tech/) website.

## CLI

### Installation

	go install github.com/saiko-tech/csgo-centrifuge/cmd/csgo-centrifuge@latest

### Usage

```
$ csgo-centrifuge --help
NAME:
   csgo-centrifuge - process CSGO game files in (hopefully) interesting ways

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   bsp             extract interesting data from BSP (Binary-Space-Partition - source-engine maps) files
   crc-table, crc  extract the CRC table from bin/linux64/engine_client.so
   download, dl    download a file from the steam workshop
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

#### Example

If you have installed `csgo-centrifuge`, [`cq`](https://github.com/markus-wa/cq) & [ImageMagick](https://imagemagick.org/index.php)'s `convert` you can do the following to get the correct radar image from a map_crc code.

See further down on [how to get the map_crc code](#how-to-get-the-map-crc-code).

```terminal
$ map_crc=2895852907
$ csgo_dir="SteamLibrary/steamapps/common/Counter-Strike Global Offensive"

$ crc_table=$(csgo-centrifuge crc-table --in-file $csgo_dir/bin/linux64/engine_client.so)

$ echo $crc_table | cq "(filter #(= (:map_crc %) $map_crc)) first"
{:map_crc 2895852907, :map_name "de_cache", :workshop_id 2497723828}

$ echo $crc_table | cq "(filter #(= (:map_crc %) $map_crc)) first :map_name"
de_cache

$ echo $crc_table | cq "(filter #(= (:map_crc %) $map_crc)) first :workshop_id"
2497723828

$ map_name=de_cache
$ workshop_id=2650330155

$ csgo-centrifuge download --workshop-file-id 2497723828 --out-file $map_name.bsp.zip
$ unzip $map_name.bsp.zip

$ csgo-centrifuge bsp radar-image --in-file de_cache.bsp --output-dir out
$ ls out
de_cache_radar.dds  de_cache.txt

$ convert -flip out/de_cache_radar.dds de_cache_radar.png
```

And then you get the following image `de_cache_radar.png`:

<p align="center">
   <img alt="sample output radar image" src="https://user-images.githubusercontent.com/5138316/144641388-46b1744e-01fc-48be-b5b7-065cf2e4c6cf.png" width="50%">
</p>


## Library

### Go Get

#### BSP Utils (Radar Extraction)

	go get github.com/saiko-tech/csgo-centrifuge/pkg/bsputil@latest
	
#### CRC Table Extraction

	go get github.com/saiko-tech/csgo-centrifuge/pkg/crc@latest
	
#### Steam API (Workshop Downloads)
	go get github.com/saiko-tech/csgo-centrifuge/pkg/steamapi@latest

### Usage

See [API docs](https://pkg.go.dev/github.com/saiko-tech/csgo-centrifuge/pkg?tab=doc).

### How to get the map crc code

You can get the map_crc code from demos in the net-message `msg.CSVCMsg_ServerInfo.MapCrc` using [`demoinfocs-golang`](https://github.com/markus-wa/demoinfocs-golang).

```go
package main

import (
	"fmt"
	"os"

	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/msg"
)

func main() {
	f, err := os.Open("my.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := dem.NewParser(f)
	defer p.Close()

	p.RegisterNetMessageHandler(func(info *msg.CSVCMsg_ServerInfo) {
		fmt.Println("map_crc", info.MapCrc)
	})
   
	err := p.ParseToEnd()
	if err != nil {
		panic(err)
	}
}
```

### Acknowledgements

Massive thanks to [@rogerxiii](https://github.com/rogerxiii) for the proof of concept & help along the way!
