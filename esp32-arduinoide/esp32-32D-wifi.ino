#include <WiFi.h>
#include <WiFiUdp.h>
#include <WebServer.h>
#include <FastLED.h>

#define MAX_LEDS 300

// IP ESTÁTICA
IPAddress local_IP(192,168,1,69);
IPAddress gateway(192,168,1,1);
IPAddress subnet(255,255,255,0);
//

CRGB leds[MAX_LEDS];

WiFiUDP udp;
WebServer server(80);

uint16_t NUM_LEDS = 84;
uint8_t LED_PIN = 25;
uint16_t FPS = 60;

unsigned long frameInterval = 1000 / 60;

const char* ssid     = "ssid";
const char* password = "pass";

// =====================
// WEB UI
// =====================

String htmlPage() {
    return R"rawliteral(
    <html>
    <body>
    <h2>ESP32 LED Config</h2>
    <form action="/set">
      FPS: <input name="fps"><br>
      NUM_LEDS: <input name="leds"><br>
      LED_PIN: <input name="pin"><br>
      <input type="submit" value="Apply">
    </form>
    </body>
    </html>
    )rawliteral";
}

void handleRoot() {
    server.send(200, "text/html", htmlPage());
}

void handleSet() {
    if (server.hasArg("fps")) {
        FPS = server.arg("fps").toInt();
        frameInterval = 1000 / FPS;
    }
    if (server.hasArg("leds")) {
        long requested = server.arg("leds").toInt();
        if (requested < 1) requested = 1;
        if (requested > MAX_LEDS) requested = MAX_LEDS;

        NUM_LEDS = (uint16_t)requested;
    }
    if (server.hasArg("pin")) {
        LED_PIN = server.arg("pin").toInt();
        FastLED.clear();
        FastLED.addLeds<WS2812, 25, GRB>(leds, NUM_LEDS);
    }

    server.send(200, "text/plain", "OK");
}

void TaskLED(void * parameter) {

    static uint8_t buffer[MAX_LEDS * 3];

    for (;;) {

        int packetSize = udp.parsePacket();
        if (packetSize > 0) {

            int len = udp.read(buffer, sizeof(buffer));

            if (len >= NUM_LEDS * 3) {

                for (int i = 0; i < NUM_LEDS; i++) {
                    leds[i].r = buffer[i * 3];
                    leds[i].g = buffer[i * 3 + 1];
                    leds[i].b = buffer[i * 3 + 2];
                }

                FastLED.show();
            }
        }

        vTaskDelay(1); // ceder CPU
    }
}

// =====================
// SETUP
// =====================

void setup() {
    Serial.begin(115200);
    WiFi.config(local_IP, gateway, subnet);
    WiFi.begin(ssid, password);
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
    }

    FastLED.addLeds<WS2812, 25, GRB>(leds, NUM_LEDS);
    FastLED.setBrightness(255);

    udp.begin(7770);

    server.begin();

    // Crear tarea en Core 1
    xTaskCreatePinnedToCore(
        TaskLED,
        "TaskLED",
        10000,
        NULL,
        2,      // prioridad
        NULL,
        1       // Core 1
    );
}

// =====================
// LOOP
// =====================

void loop() {
    server.handleClient();
}