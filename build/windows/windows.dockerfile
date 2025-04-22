# install all dependencies 
FROM golang:1.23-windowsservercore AS build_env

# wails
RUN go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest

# vc redist
ADD https://aka.ms/vs/17/release/vc_redist.x64.exe /deps/vc_redist.x64.exe
RUN /deps/vc_redist.x64.exe /install /quiet /norestart

# node.js
ADD https://nodejs.org/dist/v22.14.0/node-v22.14.0-x64.msi /deps/node.msi
RUN msiexec.exe /i C:\deps\node.msi /qn

# winfsp
ADD https://github.com/winfsp/winfsp/releases/download/v2.0/winfsp-2.0.23075.msi /deps/winfsp.msi
RUN msiexec.exe /i C:\deps\winfsp.msi /qn ADDLOCAL="F.Developer"

# winlibs
ADD https://github.com/brechtsanders/winlibs_mingw/releases/download/14.2.0posix-19.1.7-12.0.0-msvcrt-r3/winlibs-x86_64-posix-seh-gcc-14.2.0-mingw-w64msvcrt-12.0.0-r3.zip /deps/winlibs.zip
RUN powershell.exe -Command Expand-Archive -LiteralPath C:\deps\winlibs.zip -DestinationPath C:\winlibs
RUN setx PATH \"$env:PATH;C:\winlibs\mingw64\bin\"

WORKDIR /tagger
