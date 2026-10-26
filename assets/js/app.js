import "@hotwired/turbo";
import "./velle/main.js";
import { Application } from "@hotwired/stimulus";
import ButtonController from "./controllers/bridge/button_controller.js";
import RefreshController from "./controllers/refresh_controller.js";

const application = Application.start();
application.register("bridge--button", ButtonController);
application.register("refresh", RefreshController);
