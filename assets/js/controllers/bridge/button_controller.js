import { BridgeComponent } from "@hotwired/hotwire-native-bridge"

export default class extends BridgeComponent {
  static component = "button"

  connect() {
    super.connect()

    if (this.enabled) {
      this.send("connect", { title: this.bridgeElement.title }, () => {
        this.element.click()
      })
    }
  }
}
