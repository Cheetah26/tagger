# About

Store and organize files with a tagging system.

# Building

Linux:
- Execute `wails build`

Windows:
- Install [WinFSP](https://winfsp.dev/) w/ "Developer" features selected
- Install a C compilation toolchain (e.g. MinGW-w64 via [WinLibs](https://winlibs.com/))
- Add FUSE libary to search path with `go env -w CGO_CFLAGS="-O2 -g -I 'C:\Program Files (x86)\WinFsp\inc\fuse'"`
- Proceed normally with `wails build`


# Development

Follow the build steps to get your system setup, then execute `wails dev` to start automatic rebuils on file changes.