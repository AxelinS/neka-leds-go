cmake_minimum_required(VERSION 3.10)
project(dxgi_capture VERSION 1.0 LANGUAGES CXX)

if(MSVC)
  set(CMAKE_WINDOWS_EXPORT_ALL_SYMBOLS ON)
endif()

add_library(dxgi_capture SHARED dxgi_capture.cpp)

target_compile_features(dxgi_capture PRIVATE cxx_std_17)

find_library(D3D11_LIB d3d11)
find_library(DXGI_LIB dxgi)

target_link_libraries(dxgi_capture PRIVATE ${D3D11_LIB} ${DXGI_LIB})

set_target_properties(dxgi_capture PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_SOURCE_DIR}/.."
)
