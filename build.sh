#!/bin/bash

### Linux ###

go build .

### Windows ###

mkdir -p build

CGO_ENABLED=1 \
	CGG_FLAGS="-j32" \
	PKG_CONFIG_PATH=/usr/x86_64-w64-mingw32/lib/pkgconfig \
	CC="x86_64-w64-mingw32-gcc" \
	GOOS=windows \
	GOARCH=amd64 \
	go build -v -o build/glfractal.exe .

rsync /usr/x86_64-w64-mingw32/bin/libcairo-2.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libcairo-gobject-2.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgdk-3-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgtk-3-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libfontconfig-1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgdk_pixbuf-2.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgio-2.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libglib-2.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgobject-2.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpango-1.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libintl-8.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpangocairo-1.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libssp-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgmodule-2.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libjpeg-8.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpng16-16.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgcc_s_seh-1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libfreetype-6.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libexpat-1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libtiff-6.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpixman-1-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/zlib1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpcre2-8-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libepoxy-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libfribidi-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpangowin32-1.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libffi-8.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libiconv-2.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libpangoft2-1.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libharfbuzz-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libthai-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libwinpthread-1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libbz2-1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libbrotlidec.dll build/
rsync /usr/x86_64-w64-mingw32/bin/liblzma-5.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libatk-1.0-0.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libstdc++-6.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libgraphite2.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libdatrie-1.dll build/
rsync /usr/x86_64-w64-mingw32/bin/libbrotlicommon.dll build/

zip -jq glfractal.zip build/*
