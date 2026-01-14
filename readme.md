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

- Ejecuta las pruebas Go (en la raíz del repo):
go test ./src/win -run TestDXGIInitAndCapture -v
(Si no aparece la DLL en la carpeta del binario, copia src\win\dxgi_capture.dll a la raíz o junto al ejecutable).

- Ejecutar la app
Compila/executa tu aplicación como de costumbre (la captura elegirá DXGI automáticamente si la DLL y la inicialización funcionan):
```
go build -ldflags "-H=windowsgui" .
```