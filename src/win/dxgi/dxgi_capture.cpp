#include <windows.h>
#include <d3d11.h>
#include <dxgi1_2.h>
#include <wrl/client.h>
#include <string>

using Microsoft::WRL::ComPtr;

static ComPtr<ID3D11Device> g_device;
static ComPtr<ID3D11DeviceContext> g_context;
static ComPtr<IDXGIOutputDuplication> g_duplication;
static ComPtr<ID3D11Texture2D> g_staging;
static int g_width = 0;
static int g_height = 0;

static void DebugLog(const char *msg) {
    OutputDebugStringA(msg);
}

extern "C" __declspec(dllexport) bool __stdcall DXGI_Init(int width, int height) {
    if (g_duplication) return true;

    UINT createDeviceFlags = 0;
    D3D_FEATURE_LEVEL featureLevels[] = { D3D_FEATURE_LEVEL_11_0 };
    D3D_FEATURE_LEVEL createdLevel;

    HRESULT hr = D3D11CreateDevice(
        nullptr,
        D3D_DRIVER_TYPE_HARDWARE,
        nullptr,
        createDeviceFlags,
        featureLevels,
        1,
        D3D11_SDK_VERSION,
        &g_device,
        &createdLevel,
        &g_context
    );

    if (FAILED(hr)) {
        DebugLog("DXGI_Init: D3D11CreateDevice failed\n");
        return false;
    }

    ComPtr<IDXGIDevice> dxgiDevice;
    hr = g_device.As(&dxgiDevice);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: As<IDXGIDevice> failed\n");
        return false;
    }

    ComPtr<IDXGIAdapter> adapter;
    hr = dxgiDevice->GetAdapter(&adapter);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: GetAdapter failed\n");
        return false;
    }

    ComPtr<IDXGIOutput> output;
    hr = adapter->EnumOutputs(0, &output);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: EnumOutputs failed\n");
        return false;
    }

    ComPtr<IDXGIOutput1> output1;
    hr = output.As(&output1);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: As<IDXGIOutput1> failed\n");
        return false;
    }

    hr = output1->DuplicateOutput(g_device.Get(), &g_duplication);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: DuplicateOutput failed\n");
        return false;
    }

    // Create a staging texture we'll copy into for CPU-read
    D3D11_TEXTURE2D_DESC desc = {};
    desc.Width = width;
    desc.Height = height;
    desc.MipLevels = 1;
    desc.ArraySize = 1;
    desc.Format = DXGI_FORMAT_B8G8R8A8_UNORM;
    desc.SampleDesc.Count = 1;
    desc.Usage = D3D11_USAGE_STAGING;
    desc.BindFlags = 0;
    desc.CPUAccessFlags = D3D11_CPU_ACCESS_READ;

    hr = g_device->CreateTexture2D(&desc, nullptr, &g_staging);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: CreateTexture2D staging failed\n");
        // still keep duplication but fail
        g_duplication.Reset();
        return false;
    }

    g_width = width;
    g_height = height;

    return true;
}

extern "C" __declspec(dllexport) int __stdcall DXGI_GetFrame(unsigned char* dest, int destLen) {
    if (!g_duplication) return 0;
    if (!dest) return 0;

    IDXGIResource* desktopResource = nullptr;
    DXGI_OUTDUPL_FRAME_INFO frameInfo = {};
    HRESULT hr = g_duplication->AcquireNextFrame(500, &frameInfo, &desktopResource);
    if (FAILED(hr)) {
        if (hr == DXGI_ERROR_WAIT_TIMEOUT) {
            return 0;
        }
        char msg[256];
        sprintf(msg, "DXGI_GetFrame: AcquireNextFrame failed hr=0x%08x\n", (unsigned)hr);
        DebugLog(msg);
        // If duplication was lost, ask caller to reinit (-2)
        if (hr == DXGI_ERROR_ACCESS_LOST || hr == DXGI_ERROR_ACCESS_DENIED || hr == DXGI_ERROR_SESSION_DISCONNECTED) {
            // release duplication - caller can call DXGI_Init again
            g_duplication.Reset();
            sprintf(msg, "DXGI_GetFrame: duplication lost, requesting reinit\n");
            DebugLog(msg);
            return -2;
        }
        return -1;
    }

    ComPtr<ID3D11Texture2D> acquiredTex;
    hr = desktopResource->QueryInterface(__uuidof(ID3D11Texture2D), (void**)acquiredTex.GetAddressOf());
    desktopResource->Release();
    if (FAILED(hr)) {
        DebugLog("DXGI_GetFrame: QueryInterface ID3D11Texture2D failed\n");
        g_duplication->ReleaseFrame();
        return 0;
    }

    g_context->CopyResource(g_staging.Get(), acquiredTex.Get());

    D3D11_MAPPED_SUBRESOURCE mapped;
    hr = g_context->Map(g_staging.Get(), 0, D3D11_MAP_READ, 0, &mapped);
    if (FAILED(hr)) {
        DebugLog("DXGI_GetFrame: Map failed\n");
        g_duplication->ReleaseFrame();
        return 0;
    }

    int rowSize = g_width * 4;
    int bytesCopied = 0;
    unsigned char* srcRow = reinterpret_cast<unsigned char*>(mapped.pData);
    for (int y = 0; y < g_height; y++) {
        int copyLen = rowSize;
        if (bytesCopied + copyLen > destLen) {
            copyLen = destLen - bytesCopied;
            if (copyLen <= 0) break;
        }
        memcpy(dest + bytesCopied, srcRow, copyLen);
        bytesCopied += copyLen;
        srcRow += mapped.RowPitch;
    }

    g_context->Unmap(g_staging.Get(), 0);
    g_duplication->ReleaseFrame();

    return bytesCopied;
}

extern "C" __declspec(dllexport) void __stdcall DXGI_Close() {
    g_staging.Reset();
    g_duplication.Reset();
    g_context.Reset();
    g_device.Reset();
    g_width = g_height = 0;
}
