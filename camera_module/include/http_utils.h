#pragma once

#include <WiFiClient.h>

String read_http_response(WiFiClient *client, int timeout_ms)
{
  unsigned long startTime = millis();
  while (!client->available())
  {
    if (millis() - startTime > timeout_ms)
    {
      log_e("[HTTP] Response timed out");
      client->stop();

      return "";
    }
  }
  
  String response;
  while (client->available())
  {
    response += client->readStringUntil('\r');
  }
  
  client->stop();
  return response;
}
