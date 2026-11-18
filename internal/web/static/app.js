"use strict";

const switcher = document.querySelector("[data-switcher]");
if (switcher) {
  const versions = switcher.querySelector("select");
  versions.addEventListener("change", () => switcher.requestSubmit());
  switcher.querySelector(".switch-submit").hidden = true;
}

for (const preview of document.querySelectorAll("[data-calorie-preview]")) {
  const fields = preview.querySelectorAll("[data-kcal-per-gram]");
  const result = preview.querySelector("[data-calorie-result]");
  preview.addEventListener("input", () => {
    let calories = 0;
    for (const field of fields) {
      const raw = field.value.trim().replaceAll(",", ".");
      const grams = Number(raw);
      if (!/^[\d.]+$/.test(raw) || raw.length > 16 || !Number.isFinite(grams) || grams > 1_000_000) {
        result.textContent = "—";
        return;
      }
      calories += grams * Number(field.dataset.kcalPerGram);
    }
    result.textContent = String(Math.round(calories)).replace(/\B(?=(\d{3})+(?!\d))/g, "\u202f");
  });
}
