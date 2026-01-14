DXGI capture plugin

This folder contains a native DXGI-based capture implementation to capture full-screen accelerated apps and games.

Build

- Build `dxgi_capture.dll` using CMake or MSVC as described in `src/win/dxgi/README.md`.
- Copy `dxgi_capture.dll` next to the program executable or to a directory in PATH.

Usage

- The Go wrapper `dxgi_capture.go` will try to load `dxgi_capture.dll` automatically when `NewScreenCapturer` is called.
- If the DLL is present and initialization succeeds, the DXGI backend will be used automatically; otherwise the GDI fallback is used.

Notes

- The plugin returns BGRA pixel data (4 bytes per pixel), same layout as the GDI capture.
- Some exclusive or protected content may still not be available depending on the system and GPU drivers.
