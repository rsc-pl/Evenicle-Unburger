# Building the Evenicle Unburger Mod

The repository contains the translation data and build tools. It does **not** contain the original Evenicle game files.

A normal build starts from your own installed copy of the MangaGamer English PC release, version 1.04. The public build path is implemented in Go and does not require Python or Pillow.

## Quick start: use the TUI

For normal interactive use, the TUI is the recommended frontend. It uses the exact same Go build/install pipeline and byte-identity checks as the CLI. The main screen is a color dashboard with separate environment, readiness and action panels; the selected action is highlighted and every action explains what it will do before you run it. Set `NO_COLOR=1` if you need a plain terminal display.

Linux:

```bash
./tui.sh
```

Windows:

```bat
tui.cmd
```

On the first run:

1. choose **Setup / reconfigure**;
2. select your Evenicle 1.04 installation;
3. let the wizard detect AliceTools, or select `alice` / `alice.exe`;
4. let it capture and validate the pristine game files;
5. choose **Build patch** to create the release ZIP, or **Build + install patch** to build and install it immediately.

After setup, normal use does not require manually editing `.local/config.env`. The main menu also provides **Recapture pristine base files**, **Clean generated build files** and **Setup / reconfigure**.

If you prefer commands instead, the short CLI equivalent after one-time setup is:

```bash
./build.sh install
```

```bat
build.cmd install
```

`install` performs the full build and then the guarded installation; a separate `build` command is not required first.

## Requirements

You need:

- Go 1.23 or newer;
- [AliceTools](https://github.com/nunuhara/alice-tools), with `alice` on Linux or `alice.exe` on Windows;
- a C toolchain with zlib available to Go/CGO;
- an installed, unmodified English copy of Evenicle 1.04.

CGO/zlib is used only for deterministic release-ZIP compression. The historical dev.383 release was produced with Python's `zipfile`, which uses zlib DEFLATE; the Go builder deliberately reproduces that byte layout and compressed stream instead of silently producing a merely equivalent ZIP. A build of the unchanged dev.383 source hard-checks the complete known release SHA-256 and fails if byte identity is lost.

On Linux, install your distribution's C compiler and zlib development package if they are not already present. On Windows, use a CGO-capable Go setup such as MinGW-w64/MSYS2 with zlib available to the compiler. `CGO_ENABLED=0` is intentionally rejected for release packaging because it cannot guarantee the historical byte-identical ZIP.

AliceTools remains an external dependency. The builder looks for it in this order:

1. `--alice PATH`;
2. `ALICE_BIN` from the environment or `.local/config.env`;
3. `alice` / `alice.exe` in `PATH`;
4. the project-local `.local/tools/alice` / `.local/tools/alice.exe` development fallback, if present.

A configured value containing a path separator is treated as an explicit path and fails immediately if it is wrong. Before a build starts, the executable is invoked with `--help` so a broken AliceTools installation is caught early.

## First-time setup

The easiest setup path is the interactive TUI. It uses the same Go builder as the CLI; it is only a friendlier frontend.

### Linux

```bash
./tui.sh
```

### Windows

```bat
tui.cmd
```

Use Up/Down and Enter in the main menu. The **Setup / reconfigure** wizard will:

1. ask for your Evenicle installation directory;
2. find and validate the required pristine English 1.04 files;
3. copy those files into the ignored local source cache under `.local/source/en`;
4. detect AliceTools or ask for `alice` / `alice.exe`;
5. write `.local/config.env` for you;
6. optionally build the patch immediately.

If the selected game directory is already patched by this builder, the wizard can recover the supported pristine files from verified `.unburger-backup` files instead of asking you to restore the installation first. Capture is staged and validated before replacing an existing local source cache, so choosing the wrong directory cannot silently destroy a previously valid capture.

The TUI main menu also provides **Build patch**, **Build + install patch**, **Recapture pristine base files** and **Clean generated build files**.

### Manual setup / CLI

You can still configure everything by hand. Create the ignored local configuration directory:

```text
.local/
```

Create `.local/config.env` and point it at your Evenicle installation:

```text
EVENICLE_EN_ROOT="/path/to/Evenicle"
```

Windows paths are accepted as well, for example:

```text
EVENICLE_EN_ROOT="D:\Games\Evenicle"
```

If AliceTools is not in `PATH`, add its executable path:

```text
ALICE_BIN="/path/to/alice"
```

or on Windows:

```text
ALICE_BIN="C:\Tools\AliceTools\alice.exe"
```

Capture and validate the supported base files once:

#### Linux

```bash
./build.sh capture
```

#### Windows

```bat
build.cmd capture
```

You can also pass the game directory directly:

```text
./build.sh capture --game-root /path/to/Evenicle
build.cmd capture --game-root D:\Games\Evenicle
```

The capture validates `Version.txt` and the expected SHA-256 values for `Evenicle.ain`, `Evenicle.ex` and `EvenicleCG2.afa`. Captured game files are stored under `.local/` and are ignored by Git.

A normal public build does not need the Japanese updater or the private review corpus used while the translation was being developed.

## Building

From the TUI, choose **Build patch**. The TUI calls the exact same build pipeline and byte-identity checks as the CLI.

For command-line or automated builds:

### Linux

```bash
./build.sh
```

### Windows

```bat
build.cmd
```

Both wrappers run the same Go command. You can also invoke it directly:

```text
go run ./cmd/evenicle-builder build
```

The builder will:

1. locate and invoke AliceTools before changing build output;
2. validate the captured English 1.04 base by SHA-256;
3. validate `patch/changes.tsv` and prepare the AliceTools AIN edit file;
4. generate the startup screen from `patch/startup_screen.toml` without Python/Pillow;
5. apply the managed text changes to pristine `Evenicle.ain`;
6. keep `Evenicle.ex` byte-identical to the supported base;
7. rebuild `EvenicleCG2.afa` with only the startup-screen overlay added;
8. independently dump/compare the rebuilt AIN and CG2 payload tree;
9. create the deterministic release ZIP and checksum;
10. for the unchanged dev.383 canonical inputs, enforce the known SHA-256 of every rebuilt game file and the complete release archive.

The finished files are copied to:

```text
release/Evenicle-Unburger-v<version>.zip
release/Evenicle-Unburger-v<version>.zip.sha256
```

`release/`, `output/` and `.local/` are generated locally and are not tracked.

## Installing a local build

### TUI

Launch `./tui.sh` on Linux or `tui.cmd` on Windows and choose **Build + install patch**. If setup has not been completed yet, use **Setup / reconfigure** first; the wizard writes `.local/config.env` and captures the required pristine files for you.

### CLI

With `EVENICLE_EN_ROOT` configured:

Linux:

```bash
./build.sh install
```

Windows:

```bat
build.cmd install
```

`install` always performs a fresh full build before installation, so you do not need to run `./build.sh` or `build.cmd` separately first.

The installer refuses to overwrite an unknown game state. It accepts either the supported pristine 1.04 files or an already-patched installation with verified pristine `.unburger-backup` files. The generated files are always rebuilt from the captured pristine base, never from the currently installed patch.

## Editing the patch

### Translation text

The public text authority is:

```text
patch/changes.tsv
```

Each managed entry identifies the AIN string to replace and the replacement English text. The Go builder reads this file directly.

The private review corpus used while developing the translation is not part of the public repository and is not required for a public build.

### Startup screen

The owner-editable startup copy is:

```text
patch/startup_screen.toml
```

For the unchanged dev.383 copy and version, the builder reproduces the exact runtime-tested startup WebP bytes. Those bytes are embedded in Go source as a reproducibility seed, not tracked as a generated `.webp` file. If the copy or version is deliberately changed, the Go renderer uses embedded glyph rasters and asks AliceTools to encode the generated image as WebP. Python, Pillow and installed system fonts are not required.

The glyph rasters are derived from DejaVu Sans/Serif; their license notice is included under `internal/startup/assets/`.

### Version

The patch version comes from:

```text
version.txt
```

The same value is used for the release filename, generated startup screen and version string shown by the game.

## Builder commands

```text
./tui.sh                    # interactive setup/build frontend (Linux)
tui.cmd                     # interactive setup/build frontend (Windows)
./build.sh build            # full build (default if no command is given)
./build.sh tui              # launch the same TUI through the CLI wrapper
./build.sh capture          # capture + validate pristine English 1.04 files
./build.sh render-startup   # regenerate only the startup WebP
./build.sh clean            # remove generated build/output state
./build.sh install          # full build + guarded local installation
```

Use the same command names after `build.cmd` on Windows.

Common options:

```text
--alice PATH
--game-root PATH
--root PATH
```

## Repository hygiene

Game files and machine-local working data belong under `.local/` or generated output directories. The public repository should not contain AIN, EX, AFA, executable or release archive files. Generated startup WebP files also remain outside tracked source.
