"use strict";

const switcher = document.querySelector("[data-switcher]");
if (switcher) {
  const versions = switcher.querySelector("select");
  versions.addEventListener("change", () => switcher.requestSubmit());
  switcher.querySelector(".switch-submit").hidden = true;
}
