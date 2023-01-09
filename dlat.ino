// for dlat.exe
// Measure display latency
//
// Arduino Leonardo
//
// 5V - resistor - CdS - GND
//               |_ A0

enum {
  kThreshold = 980
};

void setup() {
  Serial.begin(38400);
}

void loop() {
  static char s_OldValue = 'z';

  int v = analogRead(A0);
  char value = (v < kThreshold);
  if ( s_OldValue != value ) {
    Serial.print(value);
    s_OldValue = value;
  }
  delay(1);
}
