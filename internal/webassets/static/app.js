const optimizeForm = document.getElementById("optimize-form");
const itemsOrderedInput = document.getElementById("items-ordered");
const packSizesInput = document.getElementById("pack-sizes");
const resultSection = document.getElementById("result");
const summary = document.getElementById("summary");
const packsBody = document.getElementById("packs-body");
const errorText = document.getElementById("error");

function hideError() {
  errorText.classList.add("hidden");
  errorText.textContent = "";
}

function showError(message) {
  errorText.textContent = message;
  errorText.classList.remove("hidden");
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

function updateQueryString(itemsOrdered, packSizes) {
  const params = new URLSearchParams();
  params.set("items_ordered", String(itemsOrdered));
  params.set("pack_sizes", packSizes.join(","));

  const query = params.toString();
  const nextUrl = query
    ? `${window.location.pathname}?${query}`
    : window.location.pathname;
  window.history.replaceState({}, "", nextUrl);
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

async function runOptimization(itemsOrdered, packSizes) {
  const payload = { items_ordered: itemsOrdered, pack_sizes: packSizes };
  const data = await apiFetch("/api/optimize", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  renderResult(data);
}

optimizeForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  hideError();

  const itemsOrdered = parseItemsOrdered(itemsOrderedInput.value);
  if (itemsOrdered === null) {
    showError("items_ordered must be a positive integer.");
    return;
  }

  try {
    const packSizes = parsePackSizes(packSizesInput.value);
    await runOptimization(itemsOrdered, packSizes);
    updateQueryString(itemsOrdered, packSizes);
  } catch (err) {
    showError(err.message);
  }
});

async function initializeFromQueryString() {
  hideError();
  const params = new URLSearchParams(window.location.search);
  const queryItemsOrdered = params.get("items_ordered");
  const queryPackSizes = params.get("pack_sizes");

  if (queryPackSizes !== null) {
    packSizesInput.value = queryPackSizes;
  }

  if (queryItemsOrdered === null) {
    return;
  }

  const itemsOrdered = parseItemsOrdered(queryItemsOrdered);
  if (itemsOrdered === null) {
    showError("Invalid items_ordered query param. It must be a positive integer.");
    return;
  }

  let packSizes;
  try {
    packSizes = parsePackSizes(queryPackSizes === null ? packSizesInput.value : queryPackSizes);
  } catch (err) {
    showError(`Invalid pack_sizes query param: ${err.message}`);
    return;
  }

  itemsOrderedInput.value = String(itemsOrdered);

  try {
    await runOptimization(itemsOrdered, packSizes);
    if (queryPackSizes === null) {
      updateQueryString(itemsOrdered, packSizes);
    }
  } catch (err) {
    showError(err.message);
  }
}

initializeFromQueryString();
