const optimizeForm = document.getElementById("optimize-form");
const updatePackSizesForm = document.getElementById("update-pack-sizes-form");
const itemsOrderedInput = document.getElementById("items-ordered");
const packSizesInput = document.getElementById("pack-sizes");
const resultSection = document.getElementById("result");
const summary = document.getElementById("summary");
const packsBody = document.getElementById("packs-body");
const errorText = document.getElementById("error");
const packSizeUpdateMessage = document.getElementById("pack-size-update-message");

function hideError() {
  errorText.classList.add("hidden");
  errorText.textContent = "";
}

function showError(message) {
  errorText.textContent = message;
  errorText.classList.remove("hidden");
}

function hideUpdateMessage() {
  packSizeUpdateMessage.classList.add("hidden");
  packSizeUpdateMessage.textContent = "";
}

function showUpdateMessage(message) {
  packSizeUpdateMessage.textContent = message;
  packSizeUpdateMessage.classList.remove("hidden");
}

async function apiFetch(path, options = {}) {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Request failed");
  }

  return data;
}

function parseItemsOrdered(value) {
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return null;
  }
  return parsed;
}

function parsePackSizes(raw) {
  const value = raw.trim();
  if (value === "") {
    throw new Error("pack_sizes is required.");
  }

  const parts = value.split(",");
  const parsed = [];
  const seen = new Set();

  for (const part of parts) {
    const trimmed = part.trim();
    if (trimmed === "") {
      throw new Error("pack_sizes must be a comma-separated list of integers.");
    }

    const n = Number(trimmed);
    if (!Number.isInteger(n) || n <= 0) {
      throw new Error("pack_sizes must contain only positive integers.");
    }

    if (!seen.has(n)) {
      seen.add(n);
      parsed.push(n);
    }
  }

  if (parsed.length === 0) {
    throw new Error("pack_sizes must contain at least one positive integer.");
  }
  return parsed;
}

function updateQueryString(itemsOrdered) {
  const params = new URLSearchParams();
  if (itemsOrdered !== null) {
    params.set("items_ordered", String(itemsOrdered));
  }

  const query = params.toString();
  const nextUrl = query
    ? `${window.location.pathname}?${query}`
    : window.location.pathname;
  window.history.replaceState({}, "", nextUrl);
}

async function fetchPackSizes() {
  const data = await apiFetch("/api/pack-sizes");
  if (!Array.isArray(data.pack_sizes)) {
    throw new Error("invalid pack_sizes response");
  }

  return parsePackSizes(data.pack_sizes.join(","));
}

async function updatePackSizes(packSizes) {
  const data = await apiFetch("/api/pack-sizes", {
    method: "PUT",
    body: JSON.stringify({ pack_sizes: packSizes }),
  });
  if (!Array.isArray(data.pack_sizes)) {
    throw new Error("invalid pack_sizes response");
  }

  return parsePackSizes(data.pack_sizes.join(","));
}

function renderResult(data) {
  summary.textContent = `${data.items_ordered} ordered -> ${data.total_items} shipped in ${data.total_packs} pack(s).`;
  packsBody.innerHTML = "";

  data.packs.forEach((pack) => {
    const row = document.createElement("tr");
    row.innerHTML = `<td>${pack.size}</td><td>${pack.count}</td>`;
    packsBody.appendChild(row);
  });

  resultSection.classList.remove("hidden");
}

async function runOptimization(itemsOrdered) {
  const payload = { items_ordered: itemsOrdered };
  const data = await apiFetch("/api/optimize", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  renderResult(data);
}

optimizeForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  hideError();
  hideUpdateMessage();

  const itemsOrdered = parseItemsOrdered(itemsOrderedInput.value);
  if (itemsOrdered === null) {
    showError("items_ordered must be a positive integer.");
    return;
  }

  try {
    await runOptimization(itemsOrdered);
    updateQueryString(itemsOrdered);
  } catch (err) {
    showError(err.message);
  }
});

updatePackSizesForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  hideError();
  hideUpdateMessage();

  try {
    const packSizes = parsePackSizes(packSizesInput.value);
    const updatedPackSizes = await updatePackSizes(packSizes);

    packSizesInput.value = updatedPackSizes.join(",");

    const itemsOrdered = parseItemsOrdered(itemsOrderedInput.value);
    updateQueryString(itemsOrdered);

    showUpdateMessage("Backend pack sizes updated.");
  } catch (err) {
    showError(err.message);
  }
});

async function initializeFromQueryString() {
  hideError();
  hideUpdateMessage();
  const params = new URLSearchParams(window.location.search);
  const queryItemsOrdered = params.get("items_ordered");

  let packSizes;
  try {
    packSizes = await fetchPackSizes();
  } catch (err) {
    showError(`Unable to load pack_sizes: ${err.message}`);
    return;
  }

  packSizesInput.value = packSizes.join(",");

  if (queryItemsOrdered === null) {
    updateQueryString(null);
    return;
  }

  const itemsOrdered = parseItemsOrdered(queryItemsOrdered);
  if (itemsOrdered === null) {
    showError("Invalid items_ordered query param. It must be a positive integer.");
    return;
  }

  itemsOrderedInput.value = String(itemsOrdered);

  try {
    await runOptimization(itemsOrdered);
    updateQueryString(itemsOrdered);
  } catch (err) {
    showError(err.message);
  }
}

initializeFromQueryString();
