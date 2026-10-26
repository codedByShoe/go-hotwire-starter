import HotwireNative
import UIKit

final class ButtonComponent: BridgeComponent {
    override class var name: String { "button" }

    override func onReceive(message: Message) {
        guard let viewController = delegate?.destination as? UIViewController,
              let data: MessageData = message.data() else { return }

        let action = UIAction { [weak self] _ in
            self?.reply(to: "connect")
        }
        viewController.navigationItem.rightBarButtonItem = UIBarButtonItem(
            title: data.title,
            primaryAction: action
        )
    }
}

private extension ButtonComponent {
    struct MessageData: Decodable {
        let title: String
    }
}
