#include <windows.h>
#include <d3d11.h>
#include <dxgi1_2.h>
#include <wrl/client.h>
#include <string>

using Microsoft::WRL::ComPtr;

// Scope guard para garantizar ReleaseFrame en salidas
struct FrameGuard {
    IDXGIOutputDuplication* dup;
    bool shouldRelease;
    
    FrameGuard(IDXGIOutputDuplication* d) : dup(d), shouldRelease(true) {}
    ~FrameGuard() {
        if (shouldRelease && dup) {
            dup->ReleaseFrame();
        }
    }
    void cancel() { shouldRelease = false; }
};

static ComPtr<ID3D11Device> g_device;
static ComPtr<ID3D11DeviceContext> g_context;
static ComPtr<IDXGIOutputDuplication> g_duplication;
static ComPtr<ID3D11Texture2D> g_staging;
static int g_width = 0;
static int g_height = 0;

static void DebugLog(const char *msg) {
    OutputDebugStringA(msg);
}

// Forward declaration
extern "C" __declspec(dllexport) void __stdcall DXGI_Close();

extern "C" __declspec(dllexport) bool __stdcall DXGI_Init() {
    // Cleanup any previous state
    DXGI_Close();

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
        DXGI_Close();
        return false;
    }

    ComPtr<IDXGIAdapter> adapter;
    hr = dxgiDevice->GetAdapter(&adapter);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: GetAdapter failed\n");
        DXGI_Close();
        return false;
    }

    ComPtr<IDXGIOutput> output;
    hr = adapter->EnumOutputs(0, &output);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: EnumOutputs failed\n");
        DXGI_Close();
        return false;
    }

    ComPtr<IDXGIOutput1> output1;
    hr = output.As(&output1);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: As<IDXGIOutput1> failed\n");
        DXGI_Close();
        return false;
    }

    hr = output1->DuplicateOutput(g_device.Get(), &g_duplication);
    if (FAILED(hr)) {
        DebugLog("DXGI_Init: DuplicateOutput failed\n");
        DXGI_Close();
        return false;
    }

    // Get actual monitor resolution from DXGI
    DXGI_OUTDUPL_DESC duplDesc;
    g_duplication->GetDesc(&duplDesc);
    int width = duplDesc.ModeDesc.Width;
    int height = duplDesc.ModeDesc.Height;

    char msg[256];
    snprintf(msg, sizeof(msg), "DXGI_Init: Got monitor resolution %dx%d\n", width, height);
    DebugLog(msg);

    // Create a staging texture with actual monitor size
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
        DXGI_Close();
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
            return 0; // No new frame - timeout
        }
        char msg[256];
        snprintf(msg, sizeof(msg), "DXGI_GetFrame: AcquireNextFrame failed hr=0x%08x\n", (unsigned)hr);
        DebugLog(msg);
        
        // If duplication was lost, request reinit (-2)
        if (hr == DXGI_ERROR_ACCESS_LOST || hr == DXGI_ERROR_ACCESS_DENIED || hr == DXGI_ERROR_SESSION_DISCONNECTED) {
            g_duplication.Reset();
            DebugLog("DXGI_GetFrame: duplication lost, requesting reinit\n");
            return -2;
        }
        return -1; // Other error
    }

    // Guard: ensure ReleaseFrame is called even if something fails
    FrameGuard guard(g_duplication.Get());
    
    ComPtr<ID3D11Texture2D> acquiredTex;
    hr = desktopResource->QueryInterface(__uuidof(ID3D11Texture2D), (void**)acquiredTex.GetAddressOf());
    desktopResource->Release();
    
    if (FAILED(hr)) {
        DebugLog("DXGI_GetFrame: QueryInterface ID3D11Texture2D failed\n");
        return 0; // Guard will release frame
    }

    // Ensure acquired texture has same size as staging. If not, request reinit.
    D3D11_TEXTURE2D_DESC texDesc;
    acquiredTex->GetDesc(&texDesc);
    if ((int)texDesc.Width != g_width || (int)texDesc.Height != g_height) {
        DebugLog("DXGI_GetFrame: Resolution changed, need reinit\n");
        return -2; // Duplication lost / size changed - caller should reinit
    }

    // Copy into staging (may fail if sizes mismatch)
    g_context->CopyResource(g_staging.Get(), acquiredTex.Get());

    D3D11_MAPPED_SUBRESOURCE mapped;
    hr = g_context->Map(g_staging.Get(), 0, D3D11_MAP_READ, 0, &mapped);
    if (FAILED(hr)) {
        DebugLog("DXGI_GetFrame: Map failed\n");
        return 0; // Guard will release frame
    }

    // Verify buffer is large enough
    int rowSize = g_width * 4;
    int requiredBytes = g_height * rowSize;
    if (destLen < requiredBytes) {
        DebugLog("DXGI_GetFrame: Buffer too small\n");
        g_context->Unmap(g_staging.Get(), 0);
        return -3; // Buffer too small
    }

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

    return bytesCopied; // Positive = frame copied
}

extern "C" __declspec(dllexport) void __stdcall DXGI_Close() {
    g_staging.Reset();
    g_duplication.Reset();
    g_context.Reset();
    g_device.Reset();
    g_width = g_height = 0;
}

extern "C" __declspec(dllexport) void __stdcall DXGI_GetSize(int* w, int* h) {
    if (w) *w = g_width;
    if (h) *h = g_height;
}

extern "C" __declspec(dllexport) bool __stdcall DXGI_IsAlive() {
    return (g_device.Get() != nullptr) && (g_context.Get() != nullptr) && (g_duplication.Get() != nullptr) && (g_staging.Get() != nullptr);
}
