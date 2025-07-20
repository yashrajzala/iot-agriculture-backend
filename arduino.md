#include <WiFi.h>           // Standard library for ESP32 Wi-Fi functionality
#include <PubSubClient.h>   // MQTT client library by Nick O'Leary for reliable communication
#include <esp_task_wdt.h>   // Library for ESP32's hardware Watchdog Timer (WDT)
#include <Preferences.h>    // ESP32 NVS (Non-Volatile Storage) library
#include <ArduinoJson.h>    // For robust JSON payload construction
#include <esp_system.h>     // For esp_random() for better random seed

// --- GLOBAL CONFIGURATION SETTINGS ---
// All user-modifiable parameters are defined here for easy access and modification.

// Wi-Fi Network Credentials (These will be loaded from NVS if set, otherwise defaults)
const char* DEFAULT_WIFI_SSID = "TP-Link_129A";
const char* DEFAULT_WIFI_PASSWORD = "37633585";

// MQTT Broker Connection Details (These will be loaded from NVS if set, otherwise defaults)
// CRITICAL: This MUST be the static IP of your Windows 10 PC running Mosquitto
const char* DEFAULT_MQTT_BROKER_HOST = "192.168.20.1";
const int MQTT_BROKER_PORT = 1883; // Standard MQTT port

// MQTT Authentication Credentials (These will be loaded from NVS if set, otherwise defaults)
// NOTE: These constants are KEPT for NVS functionality.
// We are currently using anonymous access on Mosquitto for debugging.
const char* DEFAULT_MQTT_USERNAME = "cresla";
const char* DEFAULT_MQTT_PASSWORD = "cresla123.";

// --- PHYSICAL ESP32 Client ID ---
// This is the actual MQTT client ID for THIS physical ESP32 board.
// It remains constant.
const char* PHYSICAL_MQTT_CLIENT_ID = "ESP32Simulator_Main";

// --- VIRTUAL NODE CONFIGURATION ---
// Define the greenhouse ID (common for all simulated nodes)
const char* VIRTUAL_GREENHOUSE_ID = "GH1";

// Define the IDs for your 5 virtual nodes
const int NUM_VIRTUAL_NODES = 5;
const char* VIRTUAL_NODE_IDS[NUM_VIRTUAL_NODES] = {
    "Node01",
    "Node02",
    "Node03",
    "Node04",
    "Node05"
};

// --- MQTT Topic Definitions (Dynamically generated based on virtual node IDs) ---
// These are formats; actual topics will be generated dynamically per virtual node
const char* MQTT_TOPIC_PUBLISH_FORMAT = "greenhouse/%s/node/%s/data";
// const char* MQTT_TOPIC_SUBSCRIBE_FORMAT = "greenhouse/%s/node/%s/commands"; // Removed: No longer subscribing to commands
const char* MQTT_LWT_TOPIC_FORMAT = "greenhouse/%s/node/%s/status"; // LWT for each virtual node

// MQTT Last Will and Testament (LWT) Configuration for the PHYSICAL client
// This LWT will be published if the physical ESP32 disconnects unexpectedly.
const char* PHYSICAL_MQTT_LWT_TOPIC = "simulator/status";
const char* PHYSICAL_MQTT_LWT_MESSAGE = "offline";
const int PHYSICAL_MQTT_LWT_QOS = 1;
const bool PHYSICAL_MQTT_LWT_RETAIN = true;

// --- TIMING CONFIGURATION ---
// Interval for publishing data for ALL virtual nodes (e.g., every 2 seconds, all 5 nodes send data)
const long PUBLISH_ALL_NODES_INTERVAL_MS = 2000;

// Non-blocking retry intervals with Exponential Backoff
const long MIN_RETRY_INTERVAL_MS = 1000;    // 1 second initial retry
const long MAX_RETRY_INTERVAL_MS = 60000; // 60 seconds max retry delay
const float BACKOFF_FACTOR = 1.5;           // Factor to multiply retry delay by
long currentWifiRetryInterval = MIN_RETRY_INTERVAL_MS;
long currentMqttRetryInterval = MIN_RETRY_INTERVAL_MS;

// --- HARDWARE CONFIGURATION ---
const int LED_PIN = 2; // Onboard LED pin for status indication

// Watchdog Timer Timeout (in seconds)
const int WDT_TIMEOUT_SECONDS = 10;

// --- DEBUG MODE ---
const bool DEBUG_MODE = true;

// --- END GLOBAL CONFIGURATION SETTINGS ---


// --- GLOBAL OBJECTS & VARIABLES ---

WiFiClient wifiClient;
PubSubClient mqttClient(wifiClient);
Preferences preferences; // NVS Preferences object

// Actual credentials loaded from NVS
char wifiSsid[64];
char wifiPassword[64];
char mqttBrokerHost[64];
char mqttUsername[32]; // Kept for NVS functionality, but not used in connect for anonymous mode
char mqttPassword[32]; // Kept for NVS functionality, but not used in connect for anonymous mode

// Dynamic topic strings (mqttSubscribeTopic is removed as it's no longer needed)
char mqttPublishTopic[128]; // Increased buffer size for longer topics
char mqttLwtTopic[128];

long lastPublishAllNodesTime = 0; // Tracks when all nodes last published
long lastWifiAttemptTime = 0;
long lastMqttAttemptTime = 0;

unsigned long lastLedToggleTime = 0;
bool ledState = LOW;
int ledBlinkRate = 0;

// ArduinoJson document for payload construction (size based on expected JSON)
// Max size for Node01-04 (10 fields + overhead) approx 256-300 bytes
// Max size for Node05 (5 fields + overhead) approx 128-150 bytes
// Using 512 for safety.
StaticJsonDocument<512> jsonDoc;
char jsonPayloadBuffer[512];        // Buffer for serializing JSON to


// --- DEBUG MACRO ---
#define DEBUG_PRINT(x) do { if (DEBUG_MODE) { Serial.print(x); } } while (0)
#define DEBUG_PRINTLN(x) do { if (DEBUG_MODE) { Serial.println(x); } } while (0)


// --- FUNCTION PROTOTYPES ---
// Forward declarations for functions defined later in the file
void initializeHardware();
void loadCredentialsFromNVS();
void saveCredentialsToNVS(const char* ssid, const char* pass, const char* broker, const char* user, const char* mqttPass);
void setupWiFi();
void handleWiFiConnection();
void reconnectMQTT();
// void mqttCallback(char* topic, byte* payload, unsigned int length); // Removed: No longer receiving commands
void publishVirtualNodeData(const char* greenhouseId, const char* nodeId); // Function for single virtual node
void publishAllVirtualNodesData(); // Function to loop through all virtual nodes
void handleMQTTLoop();
void setLed(int state);
void setLedBlinkRate(int rateMs);
void updateLedStatus();
const char* getMqttConnectErrorString(int state);


// --- FUNCTION IMPLEMENTATIONS ---

/**
 * @brief Initializes core hardware components (Serial, LED, Watchdog).
 */
void initializeHardware() {
  Serial.begin(115200);
  DEBUG_PRINTLN("\nSerial communication initialized.");

  pinMode(LED_PIN, OUTPUT);
  setLed(LOW);
  DEBUG_PRINTLN("LED pin initialized.");

  // Seed the random number generator using the ESP32's hardware TRNG
  randomSeed(esp_random());
  DEBUG_PRINTLN("Random number generator seeded with TRNG.");

  // Initialize the Watchdog Timer (WDT).
  esp_task_wdt_config_t wdt_config = {
    .timeout_ms = WDT_TIMEOUT_SECONDS * 1000,
    .trigger_panic = true
  };
  esp_task_wdt_init(&wdt_config);
  esp_task_wdt_add(NULL);
  DEBUG_PRINTLN("Watchdog Timer initialized.");
}

/**
 * @brief Loads Wi-Fi and MQTT credentials from NVS.
 * If not found, uses default values and prompts to save.
 */
void loadCredentialsFromNVS() {
  // Check if NVS preferences can be opened
  if (!preferences.begin("mqtt_config", false)) {
      DEBUG_PRINTLN("ERROR: Failed to open NVS preferences! Using default credentials.");
      // Fallback to default if NVS fails
      strncpy(wifiSsid, DEFAULT_WIFI_SSID, sizeof(wifiSsid) - 1);
      wifiSsid[sizeof(wifiSsid) - 1] = '\0';
      strncpy(wifiPassword, DEFAULT_WIFI_PASSWORD, sizeof(wifiPassword) - 1);
      wifiPassword[sizeof(wifiPassword) - 1] = '\0';
      strncpy(mqttBrokerHost, DEFAULT_MQTT_BROKER_HOST, sizeof(mqttBrokerHost) - 1);
      mqttBrokerHost[sizeof(mqttBrokerHost) - 1] = '\0';
      strncpy(mqttUsername, DEFAULT_MQTT_USERNAME, sizeof(mqttUsername) - 1);
      mqttUsername[sizeof(mqttUsername) - 1] = '\0';
      strncpy(mqttPassword, DEFAULT_MQTT_PASSWORD, sizeof(mqttPassword) - 1);
      mqttPassword[sizeof(mqttPassword) - 1] = '\0';
      return; // Exit function if NVS failed
  }

  // Load WiFi SSID
  if (preferences.isKey("wifi_ssid")) {
    preferences.getString("wifi_ssid", wifiSsid, sizeof(wifiSsid));
    DEBUG_PRINT("Loaded WiFi SSID: "); DEBUG_PRINTLN(wifiSsid);
  } else {
    strncpy(wifiSsid, DEFAULT_WIFI_SSID, sizeof(wifiSsid) - 1); // -1 for null terminator
    wifiSsid[sizeof(wifiSsid) - 1] = '\0'; // Ensure null termination
    DEBUG_PRINTLN("Using default WiFi SSID. Please save if different.");
  }

  // Load WiFi Password
  if (preferences.isKey("wifi_pass")) {
    preferences.getString("wifi_pass", wifiPassword, sizeof(wifiPassword));
    DEBUG_PRINTLN("Loaded WiFi Password."); // Don't print password for security
  } else {
    strncpy(wifiPassword, DEFAULT_WIFI_PASSWORD, sizeof(wifiPassword) - 1);
    wifiPassword[sizeof(wifiPassword) - 1] = '\0';
    DEBUG_PRINTLN("Using default WiFi Password. Please save if different.");
  }

  // Load MQTT Broker Host
  if (preferences.isKey("mqtt_host")) {
    preferences.getString("mqtt_host", mqttBrokerHost, sizeof(mqttBrokerHost));
    DEBUG_PRINT("Loaded MQTT Broker Host: "); DEBUG_PRINTLN(mqttBrokerHost);
  } else {
    strncpy(mqttBrokerHost, DEFAULT_MQTT_BROKER_HOST, sizeof(mqttBrokerHost) - 1);
    mqttBrokerHost[sizeof(mqttBrokerHost) - 1] = '\0';
    DEBUG_PRINTLN("Using default MQTT Broker Host. Please save if different.");
  }

  // Load MQTT Username (kept for NVS persistence)
  if (preferences.isKey("mqtt_user")) {
    preferences.getString("mqtt_user", mqttUsername, sizeof(mqttUsername));
    DEBUG_PRINT("Loaded MQTT Username: "); DEBUG_PRINTLN(mqttUsername);
  } else {
    strncpy(mqttUsername, DEFAULT_MQTT_USERNAME, sizeof(mqttUsername) - 1);
    mqttUsername[sizeof(mqttUsername) - 1] = '\0';
    DEBUG_PRINTLN("Using default MQTT Username. Please save if different.");
  }

  // Load MQTT Password (kept for NVS persistence)
  if (preferences.isKey("mqtt_pass")) {
    preferences.getString("mqtt_pass", mqttPassword, sizeof(mqttPassword));
    DEBUG_PRINTLN("Loaded MQTT Password."); // Don't print password for security
  } else {
    strncpy(mqttPassword, DEFAULT_MQTT_PASSWORD, sizeof(mqttPassword) - 1);
    mqttPassword[sizeof(mqttPassword) - 1] = '\0';
    DEBUG_PRINTLN("Using default MQTT Password. Please save if different.");
  }

  preferences.end(); // Close NVS

  // Instructions to save credentials via Serial (e.g., for initial setup)
  DEBUG_PRINTLN("\n--- NVS Configuration Instructions ---");
  DEBUG_PRINTLN("To save new credentials, type commands in Serial Monitor (115200 baud):");
  DEBUG_PRINTLN("    save_wifi <SSID> <PASSWORD>");
  DEBUG_PRINTLN("    save_mqtt <BROKER_IP> <USERNAME> <PASSWORD>");
  DEBUG_PRINTLN("    clear_nvs_and_restart");
  DEBUG_PRINTLN("Example: save_wifi MyNet MyPass123");
  DEBUG_PRINTLN("Example: save_mqtt 192.168.1.10 brokeruser brokerpass");
  DEBUG_PRINTLN("------------------------------------");
}

/**
 * @brief Saves Wi-Fi and MQTT credentials to NVS.
 * @param ssid Wi-Fi SSID.
 * @param pass Wi-Fi password.
 * @param broker MQTT broker host.
 * @param user MQTT username.
 * @param mqttPass MQTT password.
*/
void saveCredentialsToNVS(const char* ssid, const char* pass, const char* broker, const char* user, const char* mqttPass) {
  preferences.begin("mqtt_config", false);
  preferences.putString("wifi_ssid", ssid);
  preferences.putString("wifi_pass", pass);
  preferences.putString("mqtt_host", broker);
  preferences.putString("mqtt_user", user);
  preferences.putString("mqtt_pass", mqttPass);
  preferences.end();
  DEBUG_PRINTLN("Credentials saved to NVS.");
  ESP.restart(); // Restart to apply new credentials
}

/**
 * @brief Initiates the Wi-Fi connection process.
 * This function is non-blocking; the connection status is handled by handleWiFiConnection().
 */
void setupWiFi() {
  DEBUG_PRINTLN("Starting Wi-Fi connection...");
  DEBUG_PRINT("Connecting to SSID: ");
  DEBUG_PRINTLN(wifiSsid);

  setLedBlinkRate(100); // Fast blink while attempting Wi-Fi connection

  WiFi.begin(wifiSsid, wifiPassword);
  lastWifiAttemptTime = millis();
  currentWifiRetryInterval = MIN_RETRY_INTERVAL_MS; // Reset retry interval
}

/**
 * @brief Manages the Wi-Fi connection state in a non-blocking manner with exponential backoff.
 */
void handleWiFiConnection() {
  if (WiFi.status() == WL_CONNECTED) {
    // Wi-Fi is connected, set LED to slow blink for MQTT status
    if (ledBlinkRate != 500 && !mqttClient.connected()) {
        setLedBlinkRate(500);
    }
    static bool ipPrinted = false;
    if (!ipPrinted) {
      DEBUG_PRINTLN("WiFi Connected!");
      DEBUG_PRINT("ESP32 IP Address: ");
      DEBUG_PRINTLN(WiFi.localIP());
      ipPrinted = true;
    }
    return;
  }

  // If not connected, attempt reconnection with exponential backoff
  if (millis() - lastWifiAttemptTime > currentWifiRetryInterval) {
    DEBUG_PRINT("Wi-Fi not connected. Retrying in ");
    DEBUG_PRINT(currentWifiRetryInterval / 1000);
    DEBUG_PRINTLN("s...");
    WiFi.disconnect(true);
    WiFi.begin(wifiSsid, wifiPassword);
    lastWifiAttemptTime = millis();

    // Exponential backoff
    currentWifiRetryInterval *= BACKOFF_FACTOR;
    if (currentWifiRetryInterval > MAX_RETRY_INTERVAL_MS) {
      currentWifiRetryInterval = MAX_RETRY_INTERVAL_MS;
      DEBUG_PRINTLN("Max Wi-Fi retry interval reached.");
    }
    setLedBlinkRate(100);
  }
}

/**
 * @brief Attempts to reconnect to the MQTT broker if the connection is lost.
 * Includes robust retry logic and status reporting, non-blocking with exponential backoff.
 */
void reconnectMQTT() {
  if (mqttClient.connected()) {
    setLedBlinkRate(0); // Solid ON when fully connected
    currentMqttRetryInterval = MIN_RETRY_INTERVAL_MS; // Reset retry interval on success
    return;
  }

  // If not connected, and enough time has passed for a retry
  if (millis() - lastMqttAttemptTime > currentMqttRetryInterval) {
    DEBUG_PRINT("Attempting MQTT connection for physical client ID: ");
    DEBUG_PRINT(PHYSICAL_MQTT_CLIENT_ID);
    DEBUG_PRINT("...");

    // Connect using the PHYSICAL_MQTT_CLIENT_ID and LWT for the physical device
    // We are currently using anonymous access on Mosquitto for debugging.
    // If you enable authentication on Mosquitto, change this back to:
    // if (mqttClient.connect(PHYSICAL_MQTT_CLIENT_ID, mqttUsername, mqttPassword,
    //                        PHYSICAL_MQTT_LWT_TOPIC, PHYSICAL_MQTT_LWT_QOS, PHYSICAL_MQTT_LWT_RETAIN, PHYSICAL_MQTT_LWT_MESSAGE)) {
    if (mqttClient.connect(PHYSICAL_MQTT_CLIENT_ID,
                           PHYSICAL_MQTT_LWT_TOPIC, PHYSICAL_MQTT_LWT_QOS, PHYSICAL_MQTT_LWT_RETAIN, PHYSICAL_MQTT_LWT_MESSAGE)) {
      DEBUG_PRINTLN("MQTT Connected!");
      mqttClient.publish(PHYSICAL_MQTT_LWT_TOPIC, "online"); // Publish online status for simulator
      DEBUG_PRINTLN("Published simulator online status.");

      setLedBlinkRate(0);
    } else {
      DEBUG_PRINT("MQTT connection failed, rc=");
      DEBUG_PRINT(mqttClient.state());
      DEBUG_PRINTLN(getMqttConnectErrorString(mqttClient.state()));
    }
    lastMqttAttemptTime = millis();

    // Exponential backoff
    currentMqttRetryInterval *= BACKOFF_FACTOR;
    if (currentMqttRetryInterval > MAX_RETRY_INTERVAL_MS) {
      currentMqttRetryInterval = MAX_RETRY_INTERVAL_MS;
      DEBUG_PRINTLN("Max MQTT retry interval reached.");
    }
  }
}

/**
 * @brief Removed: Callback function for incoming MQTT messages (no longer subscribing to commands).
 * If you need to re-introduce command reception, uncomment this function and mqttClient.setCallback in setup().
 */
/*
void mqttCallback(char* topic, byte* payload, unsigned int length) {
  DEBUG_PRINT("MQTT Message received on Topic: [");
  DEBUG_PRINT(topic);
  DEBUG_PRINT("] Payload: ");

  for (unsigned int i = 0; i < length; i++) {
    DEBUG_PRINT((char)payload[i]);
  }
  DEBUG_PRINTLN("");

  // Example: Parse command for a specific virtual node
  // if (strstr(topic, "/commands") && strstr(topic, "Node01")) {
  //   DEBUG_PRINTLN("Command for Node01 received!");
  //   // Add specific logic for Node01 command here
  // }
}
*/

/**
 * @brief Generates random sensor values for a specific virtual node
 * and publishes them as a single JSON object to the MQTT broker.
 * Uses ArduinoJson for robust payload construction.
 * @param greenhouseId The ID of the virtual greenhouse.
 * @param nodeId The ID of the virtual node.
 */
void publishVirtualNodeData(const char* greenhouseId, const char* nodeId) {
  jsonDoc.clear(); // Clear previous content
  jsonDoc["greenhouse_id"] = greenhouseId;
  jsonDoc["node_id"] = nodeId;
  jsonDoc["timestamp"] = millis(); // Add a timestamp for data freshness

  // Determine which set of sensors to simulate based on nodeId
  if (strcmp(nodeId, "Node05") == 0) {
    // Node 05 specific sensors
    jsonDoc["Light_Par"] = random(0, 501); // 0 to 500
    jsonDoc["Air_Temp"] = random(0, 61);   // 0 to 60
    jsonDoc["Air_Rh"] = random(0, 101);    // 0 to 100
    jsonDoc["Rain"] = random(0, 2);        // 0 or 1
  } else {
    // Node 01-04 sensors (including the new Bag_Rh4)
    jsonDoc["Bag_Temp"] = random(0, 51);    // 0 to 50
    jsonDoc["Light_Par"] = random(0, 501);  // 0 to 500
    jsonDoc["Air_Temp"] = random(0, 61);    // 0 to 60
    jsonDoc["Air_Rh"] = random(0, 101);     // 0 to 100
    jsonDoc["Leaf_temp"] = random(0, 61);   // 0 to 60
    jsonDoc["drip_weight"] = random(0, 1001); // 0 to 1000
    jsonDoc["Bag_Rh1"] = random(0, 101);    // 0 to 100
    jsonDoc["Bag_Rh2"] = random(0, 101);    // 0 to 100
    jsonDoc["Bag_Rh3"] = random(0, 101);    // 0 to 100
    jsonDoc["Bag_Rh4"] = random(0, 101);    // 0 to 100
  }

  // Serialize JSON to the buffer
  size_t jsonSize = serializeJson(jsonDoc, jsonPayloadBuffer, sizeof(jsonPayloadBuffer));

  if (jsonSize == 0) {
      DEBUG_PRINTLN("ERROR: JSON serialization failed or buffer too small!");
      return; // Exit if serialization failed
  }

  // Construct the specific publish topic for this virtual node
  char currentPublishTopic[128];
  snprintf(currentPublishTopic, sizeof(currentPublishTopic), MQTT_TOPIC_PUBLISH_FORMAT, greenhouseId, nodeId);

  DEBUG_PRINT("Publishing sensor data to topic '");
  DEBUG_PRINT(currentPublishTopic);
  DEBUG_PRINT("': ");
  DEBUG_PRINTLN(jsonPayloadBuffer);

  mqttClient.publish(currentPublishTopic, jsonPayloadBuffer);
}

/**
 * @brief Loops through all defined virtual nodes and publishes their sensor data.
 */
void publishAllVirtualNodesData() {
  for (int i = 0; i < NUM_VIRTUAL_NODES; i++) {
    publishVirtualNodeData(VIRTUAL_GREENHOUSE_ID, VIRTUAL_NODE_IDS[i]);
  }
}


/**
 * @brief Manages the MQTT client loop, ensuring connection persistence and message processing.
 */
void handleMQTTLoop() {
  if (WiFi.status() == WL_CONNECTED) {
    reconnectMQTT();
    mqttClient.loop(); // PubSubClient's loop() must still be called even if not subscribing
  } else {
    if (mqttClient.connected()) {
        mqttClient.disconnect();
        DEBUG_PRINTLN("MQTT client disconnected due to Wi-Fi loss.");
    }
    // Set LED to off if both Wi-Fi and MQTT are down
    if(ledBlinkRate != 100) { // If not already fast blinking for Wi-Fi
      setLedBlinkRate(0); // Solid off or error state
      setLed(LOW); // Ensure LED is off if not connected
    }
  }
}

/**
 * @brief Sets the direct state of the LED.
 */
void setLed(int state) {
    digitalWrite(LED_PIN, state);
    ledState = state;
}

/**
 * @brief Sets the LED blinking rate.
 */
void setLedBlinkRate(int rateMs) {
    ledBlinkRate = rateMs;
    if (rateMs == 0) {
      setLed(HIGH); // Solid ON
    } else {
      setLed(LOW);
      lastLedToggleTime = millis();
    }
}

/**
 * @brief Updates the LED state based on its current blink rate.
 */
void updateLedStatus() {
    if (ledBlinkRate == 0) {
      return;
    }

    if (millis() - lastLedToggleTime >= ledBlinkRate) {
      ledState = !ledState;
      setLed(ledState);
      lastLedToggleTime = millis();
    }
}

/**
 * @brief Provides a human-readable string for PubSubClient connection error codes.
 */
const char* getMqttConnectErrorString(int state) {
    switch (state) {
        case -4: return "MQTT_CONNECTION_TIMEOUT (-4)";
        case -3: return "MQTT_CONNECTION_LOST (-3)";
        case -2: return "MQTT_CONNECT_FAILED (-2)";
        case -1: return "MQTT_DISCONNECTED (-1)";
        case 1:  return "MQTT_UNACCEPTABLE_PROTOCOL_VERSION (1)";
        case 2:  return "MQTT_IDENTIFIER_REJECTED (2)";
        case 3:  return "MQTT_SERVER_UNAVAILABLE (3)";
        case 4:  return "MQTT_BAD_USER_NAME_OR_PASSWORD (4)";
        case 5:  return "MQTT_NOT_AUTHORIZED (5)";
        default: return "UNKNOWN_ERROR";
    }
}


// --- ARDUINO CORE FUNCTIONS ---

/**
 * @brief Arduino setup function. Runs once after power-on or reset.
 * Initializes serial communication, hardware, loads config, and starts network connections.
 */
void setup() {
  initializeHardware(); // Initialize Serial, LED, Watchdog
  loadCredentialsFromNVS(); // Load Wi-Fi and MQTT credentials from NVS

  mqttClient.setServer(mqttBrokerHost, MQTT_BROKER_PORT);
  // mqttClient.setCallback(mqttCallback); // Removed: No longer setting a callback for incoming messages
  mqttClient.setKeepAlive(120); // Set MQTT keep alive to 120 seconds (default is 15)

  setupWiFi(); // Initiate Wi-Fi connection (non-blocking)
}

/**
 * @brief Arduino loop function. Runs continuously after setup.
 * Manages Wi-Fi, MQTT connection, watchdog feeding, and periodically publishes sensor data.
 */
void loop() {
  esp_task_wdt_reset(); // Feed the watchdog timer to prevent system reset

  // Check for NVS commands over Serial (for initial setup/updates)
  if (Serial.available()) {
    String command = Serial.readStringUntil('\n');
    command.trim(); // Remove newline/carriage return

    if (command.startsWith("save_wifi")) {
      int ssidStart = command.indexOf(' ');
      int passStart = command.indexOf(' ', ssidStart + 1);
      if (ssidStart != -1 && passStart != -1) {
        String ssidStr = command.substring(ssidStart + 1, passStart);
        String passStr = command.substring(passStart + 1);
        DEBUG_PRINTLN("Saving new WiFi credentials...");
        saveCredentialsToNVS(ssidStr.c_str(), passStr.c_str(), mqttBrokerHost, mqttUsername, mqttPassword);
      } else {
        DEBUG_PRINTLN("Invalid save_wifi command format. Use: save_wifi <SSID> <PASSWORD>");
      }
    } else if (command.startsWith("save_mqtt")) {
      int brokerStart = command.indexOf(' ');
      int userStart = command.indexOf(' ', brokerStart + 1);
      int passStart = command.indexOf(' ', userStart + 1);
      if (brokerStart != -1 && userStart != -1 && passStart != -1) {
        String brokerStr = command.substring(brokerStart + 1, userStart);
        String userStr = command.substring(userStart + 1, passStart);
        String passStr = command.substring(passStart + 1);
        DEBUG_PRINTLN("Saving new MQTT credentials...");
        saveCredentialsToNVS(wifiSsid, wifiPassword, brokerStr.c_str(), userStr.c_str(), passStr.c_str());
      } else {
        DEBUG_PRINTLN("Invalid save_mqtt command format. Use: save_mqtt <BROKER_IP> <USERNAME> <PASSWORD>");
      }
    } else if (command == "clear_nvs_and_restart") {
      preferences.begin("mqtt_config", false);
      preferences.clear();
      preferences.end();
      DEBUG_PRINTLN("NVS cleared. Restarting...");
      ESP.restart();
    }
  }

  handleWiFiConnection();
  handleMQTTLoop();
  updateLedStatus();

  long currentTime = millis();

  if (WiFi.status() == WL_CONNECTED && mqttClient.connected() &&
      currentTime - lastPublishAllNodesTime > PUBLISH_ALL_NODES_INTERVAL_MS) {
    lastPublishAllNodesTime = currentTime;
    publishAllVirtualNodesData(); // Call the new function to publish for all virtual nodes
  }
}