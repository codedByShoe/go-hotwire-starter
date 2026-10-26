import HotwireNative
import UIKit

@main
final class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {
        let localConfiguration = Bundle.main.url(forResource: "ios_v1", withExtension: "json")!
        let remoteConfiguration = AppEnvironment.baseURL.appending(path: "configurations/ios_v1.json")

        Hotwire.loadPathConfiguration(from: [
            .file(localConfiguration),
            .server(remoteConfiguration)
        ])
        Hotwire.registerBridgeComponents([ButtonComponent.self])
        return true
    }
}
