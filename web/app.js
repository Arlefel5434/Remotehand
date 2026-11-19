(() => {
  const status = document.querySelector("#status");
  const message = document.querySelector("#message");
  const form = document.querySelector("#text-form");
  const input = document.querySelector("#text-input");
  const controls = [...document.querySelectorAll("button:not(#pair-form button)")];
  const encoder = new TextEncoder();
  const mute = document.querySelector("#mute");
  const level = document.querySelector("#level");
  let muted = false;
  mute.onclick = () => send("mute", { muted: !muted });
  const slider = document.querySelector("#volume-slider");
  let confirmedLevel = 0;
  let editingVolume = false;
  slider.oninput = () => {
    editingVolume = true;
    level.textContent = slider.value + "%";
  };
  slider.onchange = () => {
    editingVolume = false;
    send("volume", { level: slider.valueAsNumber });
  };
  slider.onpointercancel = () => {
    editingVolume = false;
    slider.value = confirmedLevel;
    level.textContent = confirmedLevel + "%";
  };

  let socket;
  let pairing = new URLSearchParams(location.hash.slice(1)).get("pair");
  try {
    if (pairing) {
      sessionStorage.setItem("remotehand.pair", pairing);
      history.replaceState(null, "", location.pathname + location.search);
    } else {
      pairing = sessionStorage.getItem("remotehand.pair");
    }
  } catch { /* Keep the scanned credential in memory. */ }
  let reconnectAttempt = 0;
  const pairForm = document.querySelector("#pair-form");
  const pairLink = document.querySelector("#pair-link");
  pairForm.onsubmit = event => {
    event.preventDefault();
    let url;
    try { url = new URL(pairLink.value.trim()); } catch {}
    const token = url && new URLSearchParams(url.hash.slice(1)).get("pair");
    if (url?.origin !== location.origin || !/^[A-Za-z0-9_-]{16,128}$/.test(token || "")) {
      pairLink.setCustomValidity("Нужна ссылка из QR-кода для этого адреса ноутбука.");
      pairLink.reportValidity();
      return;
    }
    location.replace(url.origin + "/#pair=" + token);
    location.reload();
  };
  pairLink.oninput = () => pairLink.setCustomValidity("");
  let reconnectTimer;
  const pad = document.querySelector("#pad");
  let sensitivity = 1;
  function bindSetting(id, outputId, apply) {
    const control = document.getElementById(id);
    const output = document.getElementById(outputId);
    const key = "remotehand." + id;
    try {
      const saved = Number(localStorage.getItem(key));
      if (saved >= Number(control.min) && saved <= Number(control.max)) control.value = saved;
    } catch { /* Storage may be unavailable in private browsing. */ }
    function update() {
      output.value = control.value + "%";
      apply(control.valueAsNumber / 100);
    }
    update();
    control.oninput = update;
    control.onchange = () => {
      update();
      try { localStorage.setItem(key, control.value); } catch { /* Keep the current setting in memory. */ }
    };
  }
  bindSetting("sensitivity", "speed-value", value => { sensitivity = value; });
  bindSetting("pad-height", "height-value", value => pad.style.setProperty("--pad-scale", value));
  let gesture = null;
  const fingers = new Map();
  let scrollDistance = 0;
  let frame = 0;
  function resetGesture() {
    cancelAnimationFrame(frame);
    frame = 0;
    gesture = null;
    fingers.clear();
    scrollDistance = 0;
  }
  function flushMotion() {
    frame = 0;
    if (!gesture || !gesture.moved) return;
    const dx = Math.round((gesture.x - gesture.sentX) * sensitivity);
    const dy = Math.round((gesture.y - gesture.sentY) * sensitivity);
    if (dx || dy) {
      const limit = n => Math.max(-200, Math.min(200, n));
      send("move", { dx: limit(dx), dy: limit(dy) });
      gesture.sentX += dx / sensitivity;
      gesture.sentY += dy / sensitivity;
    }
  }
  function track(event) {
    gesture.x = event.clientX;
    gesture.y = event.clientY;
    if (Math.hypot(gesture.x - gesture.startX, gesture.y - gesture.startY) > 4) gesture.moved = true;
  }
  pad.onpointerdown = event => {
    if (event.button !== 0 || socket?.readyState !== WebSocket.OPEN) return;
    pad.setPointerCapture(event.pointerId);
    fingers.set(event.pointerId,event.clientY);
    if (fingers.size > 1 || !event.isPrimary) {
      cancelAnimationFrame(frame);
      frame = 0;
      gesture = null;
      scrollDistance = 0;
      return;
    }
    gesture = { id: event.pointerId, startX: event.clientX, startY: event.clientY,
      sentX: event.clientX, sentY: event.clientY, moved: false, startTime: event.timeStamp };
  };
  pad.onpointermove = event => {
    if (!fingers.has(event.pointerId)) return;
    const previousY = fingers.get(event.pointerId);
    fingers.set(event.pointerId,event.clientY);
    if (fingers.size === 2) {
      // The midpoint moves by half of each contact's displacement.
      scrollDistance += (event.clientY - previousY) / 2;
      const steps = Math.max(-8, Math.min(8, Math.trunc(scrollDistance / 16)));
      if (steps) {
        send("scroll", { steps });
        scrollDistance %= 16;
      }
      return;
    }
    if (gesture?.id !== event.pointerId) return;
    track(event);
    if (!frame && gesture.moved) frame = requestAnimationFrame(flushMotion);
  };
  pad.onpointerup = event => {
    fingers.delete(event.pointerId);
    scrollDistance = 0;
    if (gesture?.id !== event.pointerId) return;
    track(event);
    cancelAnimationFrame(frame);
    if (gesture.moved) flushMotion();
    else if (event.timeStamp - gesture.startTime <= 300) send("click", { button: "left" });
    resetGesture();
  };
  pad.onpointercancel = pad.onlostpointercapture = event => {
    if (fingers.delete(event.pointerId)) scrollDistance = 0;
    if (gesture?.id === event.pointerId) resetGesture();
  };
  document.querySelectorAll("[data-button]").forEach(button => {
    button.onclick = () => send("click", { button: button.dataset.button });
  });

  function setConnected(connected) {
    resetGesture();
    status.textContent = connected ? "Подключено" : "Нет связи";
    status.classList.toggle("connected", connected);
    controls.forEach((control) => { control.disabled = !connected; });
    mute.disabled = true;
    slider.disabled = true;
    editingVolume = false;
    level.textContent = "—";
  }

  function connect() {
    clearTimeout(reconnectTimer);
    if (!pairing || !/^[A-Za-z0-9_-]{16,128}$/.test(pairing)) {
      status.textContent = "Сканируйте QR-код на ноутбуке";
      pairForm.hidden = false;
      return;
    }
    if (socket && [WebSocket.OPEN, WebSocket.CONNECTING].includes(socket.readyState)) return;

    status.textContent = "Подключение…";
    const scheme = location.protocol === "https:" ? "wss:" : "ws:";
    socket = new WebSocket(`${scheme}//${location.host}/ws`, "remotehand." + pairing);

    socket.addEventListener("open", () => {
      reconnectAttempt = 0;
      message.textContent = "";
      setConnected(true);
      pairForm.hidden = true;
    });
    socket.addEventListener("close", () => {
      setConnected(false);
      pairForm.hidden = false;
      status.textContent = "Нет связи · после перезапуска сканируйте новый QR";
      const delay = Math.min(500 * (2 ** reconnectAttempt), 10000);
      reconnectAttempt += 1;
      reconnectTimer = setTimeout(connect, delay);
    });
    socket.addEventListener("error", () => socket.close());
    socket.addEventListener("message", (event) => {
      const response = JSON.parse(event.data);
      if (response.type === "volume_state") {
        muted = response.payload.muted;
        confirmedLevel = response.payload.level;
        slider.disabled = false;
        if (!editingVolume) {
          slider.value = confirmedLevel;
          level.textContent = confirmedLevel + "%";
        }
        mute.textContent = muted ? "Включить звук" : "Выключить звук";
        mute.disabled = false;
      }
      if (response.type === "error") {
        message.textContent = response.payload.message;
        if (response.payload.code === "audio_unavailable") slider.onpointercancel();
      }
    });
  }

  function send(type, payload) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      message.textContent = "Нет связи с ноутбуком";
      return false;
    }
    socket.send(JSON.stringify({
      version: 1,
      type,
      id: commandID(),
      payload,
    }));
    return true;
  }

  function commandID() {
    if (crypto.randomUUID) return crypto.randomUUID();
    return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
  }

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    const value = input.value;
    if (!value) return;
    if (encoder.encode(value).byteLength > 2048) {
      message.textContent = "Текст должен занимать не более 2048 байт";
      return;
    }
    if (send("text", { value })) {
      input.value = "";
      input.focus();
    }
  });

  document.querySelectorAll("[data-key]").forEach((button) => {
    button.addEventListener("click", () => send("key", { key: button.dataset.key }));
  });
  window.addEventListener("online", connect);
  setConnected(false);
  if (window.isSecureContext && "serviceWorker" in navigator) {
    navigator.serviceWorker.register("/sw.js").catch(() => {});
  }
  connect();
})();
