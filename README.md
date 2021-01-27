# Collage 

A CLI tool for creating collages from images.
The tool is very simple and merges many images into one large image.

## Features

- 🗂 Works with your filesystem. Just put all images into a folder (or subfolders)
- 🏞 The tool can process JPEG, PNG and GIF files
- 🔍 It selects a random subsample of images it files for the collage
- 🚀 Utilizes all cores of your CPU the run faster


## Installation

There are multiple options:
- Download from [releases](https://github.com/KeKsBoTer/collage/releases)
- Install with `go install github.com/KeKsBoTer/collage`
- Clone the repository and build it with `go build .`

## Usage

Example: Merge all images in `~/Pictures` into one collage:
```bash
collage -rows 5 ~/Pictures collage.png
```

All options
```bash
$ collage --help
A tools for creating photo collages

Usage:
  collage <arguments> [source directory] [output file]

Arguments:
  -height uint
        width of created image (default 1080)
  -rows uint
        number of image rows (default 5)
  -width int
        width of created image (default 1920)
```