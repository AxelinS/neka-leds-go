#include <FastLED.h>

#define LED_PIN     25
#define NUM_LEDS    84
#define BRIGHTNESS  255
#define LED_TYPE    WS2812
#define COLOR_ORDER GRB

#define HEADER 0xAA

CRGB leds[NUM_LEDS];

uint8_t frameBuffer[NUM_LEDS * 3];

void setup() {
    Serial.begin(115200);

    FastLED.addLeds<LED_TYPE, LED_PIN, COLOR_ORDER>(leds, NUM_LEDS);
    FastLED.setBrightness(BRIGHTNESS);

    Serial.println("leds-esp32-nekaleds");
}

void loop() {

    // -------- DISCOVERY (texto simple) --------
    if (Serial.available()) {
        if (Serial.peek() == 'W') {
            String cmd = Serial.readStringUntil('\n');
            cmd.trim();
            if (cmd == "WHO") {
                Serial.println("leds-esp32-nekaleds");
            }
            return;
        }
    }

    // -------- FRAME BINARIO --------
    if (Serial.available() >= 3) {

        if (Serial.read() != HEADER) return;

        uint16_t len = Serial.read() << 8;
        len |= Serial.read();

        if (len != NUM_LEDS * 3) return;

        size_t received = 0;
        while (received < len) {
            received += Serial.readBytes(
                frameBuffer + received,
                len - received
            );
        }

        for (int i = 0; i < NUM_LEDS; i++) {
            leds[i].r = frameBuffer[i * 3];
            leds[i].g = frameBuffer[i * 3 + 1];
            leds[i].b = frameBuffer[i * 3 + 2];
        }

        FastLED.show();
    }
}