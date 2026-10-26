import XCTest

final class CodedByShoeUITests: XCTestCase {
    func testNativeButtonNavigatesToAbout() {
        let app = XCUIApplication()
        app.launch()

        XCTAssertTrue(app.staticTexts["Web first, native ready."].waitForExistence(timeout: 15))
        app.buttons["About"].tap()
        XCTAssertTrue(app.staticTexts["One set of HTML routes."].waitForExistence(timeout: 15))
    }
}
