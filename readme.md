### - ES -

### Comandos
## Build
```
go build -ldflags "-H=windowsgui" .
```

## A continuación tienes una guía paso a paso para compilar dxgi_capture.dll desde tu "x64 Native Tools Command Prompt for VS 2022" y verificar que todo funcione:

- Limpiar y preparar la carpeta de build (ejecuta en el prompt x64 Native Tools)
En el directorio del proyecto del plugin:
```
cd .\go-leds\src\win\dxgi
```

- Borrar build anterior (si existe) y crear carpeta de build:
```
rmdir /s /q build
mkdir build
cd build
```

- Configurar con CMake
Ejecuta:
```
cmake .. -G "NMake Makefiles"
```

Espera la salida; debería detectarse el compilador y no debe aparecer el error "CMAKE_CXX_COMPILER not set".
Si aparece error: asegúrate de que estás en el prompt "x64 Native Tools…" y que cl está disponible (ejecuta cl para comprobar).

- Compilar
Ejecuta:
```
cmake --build . --config Release
```

Al final deberías ver que dxgi_capture.dll se ha generado en win (por la propiedad RUNTIME_OUTPUT_DIRECTORY).

- Verificar (opcional)
Comprueba que el archivo existe:
dir ..\dxgi_capture.dll

- Ejecutar la app
Compila/executa tu aplicación como de costumbre (la captura elegirá DXGI automáticamente si la DLL y la inicialización funcionan):
```
go build -ldflags "-H=windowsgui" .
```


### - EN -
### Commands
## Build
go build -ldflags "-H=windowsgui" .

## Below is a step-by-step guide to compile dxgi_capture.dll from your "x64 Native Tools Command Prompt for VS 2022" and verify that everything works:

Clean and prepare the build folder (run in the x64 Native Tools prompt)
In the plugin project directory:
```
cd .\go-leds\src\win\dxgi
```
Remove previous build (if it exists) and create build folder:
```
rmdir /s /q build
mkdir build
cd build
```
Configure with CMake
Run:
```
cmake .. -G "NMake Makefiles"
```
Wait for the output; the compiler should be detected and the error "CMAKE_CXX_COMPILER not set" should NOT appear.
If an error appears: make sure you are in the "x64 Native Tools…" prompt and that cl is available (run cl to verify).

Compile
Run:
```
cmake --build . --config Release
```
At the end, you should see that dxgi_capture.dll has been generated in win (due to the RUNTIME_OUTPUT_DIRECTORY property).

Verify (optional)
Check that the file exists:
```
dir ..\dxgi_capture.dll
```

(If the DLL does not appear in the binary folder, copy src\win\dxgi_capture.dll to the root directory or next to the executable.)

Run the app
Build/run your application as usual (capture will automatically choose DXGI if the DLL and initialization work correctly):
```
go build -ldflags "-H=windowsgui" .
```