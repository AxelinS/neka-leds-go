Build instructions (Windows)

Prerequisites:
- Visual Studio with C++ workload OR CMake + MSVC toolchain

Using CMake:

mkdir build
cd build
cmake .. -G "NMake Makefiles"        # or use Visual Studio generator
cmake --build .

This will produce `dxgi_capture.dll` in the parent directory (`src/win`).

Alternative (MSVC cl):
cl /EHsc /LD dxgi_capture.cpp /link d3d11.lib dxgi.lib /OUT:dxgi_capture.dll

Notes:
- The DLL exposes three functions:
  - bool __stdcall DXGI_Init(int width, int height)
  - int  __stdcall DXGI_GetFrame(unsigned char* dest, int destLen)
  - void __stdcall DXGI_Close()

The Go wrapper copies frames as BGRA (4 bytes per pixel) into a caller-provided buffer.
