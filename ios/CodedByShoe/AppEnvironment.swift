import Foundation

enum AppEnvironment {
    // The iOS Simulator shares the Mac's localhost. Use an HTTPS URL for devices.
    static let baseURL = URL(string: "http://localhost:3000/")!
}
