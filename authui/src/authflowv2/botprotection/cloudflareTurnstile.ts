/* global Turnstile */
import { Controller } from "@hotwired/stimulus";
import { getColorScheme } from "../../getColorScheme";
import {
  dispatchBotProtectionEventExpired,
  dispatchBotProtectionEventFailed,
  dispatchBotProtectionEventVerified,
} from "./botProtection";

function parseTheme(theme: string): Turnstile.Theme {
  switch (theme) {
    case "light":
      return "light";
    case "dark":
      return "dark";
    case "auto":
      return "auto";
    default:
      return "auto";
  }
}

export class CloudflareTurnstileController extends Controller {
  static values = {
    siteKey: { type: String },
    lang: { type: String },
  };

  static targets = ["widget"];

  declare siteKeyValue: string;
  declare langValue: string;
  declare widgetTarget: HTMLDivElement;

  connect() {
    window.turnstile.ready(() => {
      const colorScheme = getColorScheme();
      window.turnstile.render(this.widgetTarget, {
        sitekey: this.siteKeyValue,
        theme: parseTheme(colorScheme),
        language: this.langValue,
        callback: (token: string) => {
          dispatchBotProtectionEventVerified(token);
        },
        "error-callback": (err: string) => {
          dispatchBotProtectionEventFailed(err);

          return true; // return non-falsy value to tell cloudflare we handled error already
        },
        "expired-callback": (token: string) => {
          dispatchBotProtectionEventExpired(token);
        },
        "response-field": false,

        // below are default values, added for clarity
        size: "normal",
      });
    });
  }
}
