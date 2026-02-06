#include <FastLED.h>

#define LED_PIN     25
#define NUM_LEDS    84
#define BRIGHTNESS  255
#define LED_TYPE    WS2812
#define COLOR_ORDER GRB

CRGB leds[NUM_LEDS];
String input;

void setup() {
    Serial.begin(115200);
    Serial.setTimeout(30);
    
    FastLED.addLeds<LED_TYPE, LED_PIN, COLOR_ORDER>(leds, NUM_LEDS);
    FastLED.setBrightness(BRIGHTNESS);

    Serial.println("READY");
}

void loop() {
    if (!Serial.available()) return;

    input = Serial.readStringUntil('\n');
    input.trim();

    if (input == "WHO") {
        Serial.println("leds-esp32-fastled");
        return;
    }

    if (input.startsWith("RGB")) {
        int idx = 0;
        int pos = 4;

        while (pos < input.length() && idx < NUM_LEDS) {
            int r = input.substring(pos).toInt();
            pos = input.indexOf('-', pos) + 1;
            int g = input.substring(pos).toInt();
            pos = input.indexOf('-', pos) + 1;
            int b = input.substring(pos).toInt();
            pos = input.indexOf(' ', pos) + 1;

            leds[idx++] = CRGB(r, g, b);
        }

        FastLED.show();
        Serial.println("OK");
    }
}