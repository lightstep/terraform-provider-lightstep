resource "lightstep_webhook_destination" "hookbringsyouback" {
  project_name     = "bluestraveler"
  destination_name = "four"
  url = "https://yourwebhook.com"
}
