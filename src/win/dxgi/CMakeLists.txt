cmake_minimum_required(VERSION 3.10)
project(dxgi_capture VERSION 1.0 LANGUAGES CXX)

if(MSVC)
  set(CMAKE_WINDOWS_EXPORT_ALL_SYMBOLS ON)
endif()

add_library(dxgi_capture SHARED dxgi_capture.cpp)

target_compile_features(dxgi_capture PRIVATE cxx_std_17)

# Link with Direct3D and DXGI libraries
# On MSVC, these are found in the Windows SDK
target_link_libraries(dxgi_capture PRIVATE d3d11 dxgi)

# Output DLL directly to src/win/ directory
set_target_properties(dxgi_capture PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_SOURCE_DIR}/../"
    RUNTIME_OUTPUT_DIRECTORY_RELEASE "${CMAKE_SOURCE_DIR}/../"
    RUNTIME_OUTPUT_DIRECTORY_DEBUG "${CMAKE_SOURCE_DIR}/../"
)
